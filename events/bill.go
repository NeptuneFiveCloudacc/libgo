package events

import (
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/events/alerts"
	"github.com/megamsys/libgo/events/bills"
	_ "github.com/megamsys/libgo/events/bills"
	constants "github.com/megamsys/libgo/utils"
)

type Bill struct {
	piggyBanks string
	stop       chan struct{}
}

func NewBill(b map[string]string) *Bill {
	return &Bill{
		piggyBanks: b[constants.PIGGYBANKS],
	}
}

// Watches for new vms, or vms destroyed.
func (self *Bill) Watch(eventsChannel *EventChannel) error {
	self.stop = make(chan struct{})
	go func() {
		for {
			select {
			case event := <-eventsChannel.channel:
				switch {
				case event.EventAction == alerts.ONBOARD:
					err := self.OnboardFunc(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				case event.EventAction == alerts.DEDUCT:
					err := self.deduct(event)
					if err != nil {
						log.Warningf("Failed to process watch event: %v", err)
					}
				}
			case <-self.stop:
				log.Info("bill watcher exiting")
				return
			}
		}
	}()
	return nil
}

func (self *Bill) skip(k string) bool {
	return !strings.Contains(self.piggyBanks, k)
}

func (self *Bill) OnboardFunc(evt *Event) error {
	log.Infof("Event:BILL:onboard")
	for k, bp := range bills.BillProviders {
		if !self.skip(k) {
			err := bp.Onboard(&bills.BillOpts{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Bill) deduct(evt *Event) error {
	log.Infof("Event:BILL:deduct")
	for k, bp := range bills.BillProviders {
		if !self.skip(k) {
			err := bp.Deduct(&bills.BillOpts{})
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *Bill) Close() {
	if self.stop != nil {
		close(self.stop)
	}
}
