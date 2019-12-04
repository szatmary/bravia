package bravia

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// Custom Installation Services
// https://pro-bravia.sony.net/develop/integrate/ssip/overview/index.html

const (
	_cmdTimeout  = 2 * time.Second
	_dialTimeout = 5 * time.Second
	_dialRetry   = 5 * time.Second
)

// Bravia object
type Bravia struct {
	conn    net.Conn
	address string
	notify  chan *Message
	answer  chan *Message
}

// Bravia Creates a new Bravia and attempts to connect to device
func NewBravia(addr string) *Bravia {
	s := Bravia{
		address: addr,
		answer:  make(chan *Message),
	}
	err := s.reconnect()
	if nil != err {
		return nil
	}

	go func() {
		response := make([]byte, 24)
		for {
			_, err := io.ReadFull(s.conn, response)
			if err != nil {
				time.Sleep(_dialRetry)
				s.reconnect()
				continue
			}
			cmd := newMessage(string(response))
			switch cmd.CommandType {
			case Notify:
				s.sendNotification(cmd)
			case Answer:
				s.answer <- cmd
			default:
				s.reconnect()
			}
		}
	}()

	return &s
}

func (s *Bravia) sendNotification(cmd *Message) {
	if s.notify != nil {
		s.notify <- cmd
	}
}

func (s *Bravia) reconnect() error {
	s.sendNotification(newMessageFromBool(Notify, CONNECTED, false))
	conn, err := net.DialTimeout("tcp", s.address, _dialTimeout)
	if err != nil {
		return err
	}
	s.conn = conn
	s.sendNotification(newMessageFromBool(Notify, CONNECTED, true))
	return nil
}

// Close closes channles and disconnects from device
func (s *Bravia) Close() {
	s.conn.Close()
	close(s.answer)
	if s.notify != nil {
		close(s.notify)
	}
}

// MakeNotifyChan creates and returns a channle for notifications
// Use Message,Message to determine the modified state
// Possible Commnds retunred:
// POWR, Use Bool() to get the new power start
// AMUT, Use Bool() to get audio mute status
// PMUT, Use Bool() to get picture mute status
// VOLU, Use Int() to get volume
// INPT, Use Ints() to get the new input, The hirst value is intput tuep (1=HDMI,4=Component,5=Screen mirroring), The secons value is the input number selected
func (s *Bravia) MakeNotifyChan() chan *Message {
	s.notify = make(chan *Message)
	return s.notify
}

func (s *Bravia) exec(cmd *Message) (*Message, error) {
	if cmd == nil {
		return nil, errors.New("invalid command")
	}
	str := fmt.Sprintf("*S%c%s%s\n", cmd.CommandType, cmd.Command, cmd.Paramaters)
	fmt.Printf("cmd: %s\n", str)
	_, err := s.conn.Write([]byte(str))
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(_cmdTimeout):
		return nil, errors.New("timeout")
	case cmd = <-s.answer:
		return cmd, nil
	}
}

// SetPowerStatus  C POWR 0000000000000000 Standby (Off)
//                 C POWR 0000000000000001 Active (On)
//                 A POWR 0000000000000000 Success
//                 A POWE FFFFFFFFFFFFFFFF Error
func (s *Bravia) SetPowerStatus(on bool) error {
	answer, err := s.exec(newMessageFromBool(Control, POWR, on))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetPowerStatus E POWR ################
//                A POWR 0000000000000000 Standby (Off)
//                A POWR 0000000000000001 Active (On)
//                A POWR FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetPowerStatus() (bool, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, POWR, ""))
	if err != nil {
		return false, err
	}
	return answer.Bool()
}

// TogglePowerStatus C TPOW ################ Toggles the power status
//                   A TPOW 0000000000000000 Success
//                   A TPOW FFFFFFFFFFFFFFFF Error
func (s *Bravia) TogglePowerStatus() error {
	answer, err := s.exec(newMessageFromString(Control, TPOW, ""))
	if err != nil {
		return err
	}
	return answer.Error()
}

// SetAudioVolume C VOLU XXXXXXXXXXXXXXXX Set the volume value in the decimal digit pad on the left with "0" e.g.) 0000000000000029
//                A VOLU 0000000000000000 Success
//                A VOLU FFFFFFFFFFFFFFFF Error
func (s *Bravia) SetAudioVolume(vol int) error {
	answer, err := s.exec(newMessageFromInts(Control, VOLU, 0, vol))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetAudioVolume E VOLU ################ Retrieves the audio volume value
//                A VOLU XXXXXXXXXXXXXXXX Success with volume value
//                A VOLU FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetAudioVolume() (int, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, VOLU, ""))
	if err != nil {
		return 0, err
	}
	return answer.Int()
}

// SetAudioMute C AMUT 0000000000000000 Unmute
//              C AMUT 0000000000000001 Mute
//              A AMUT 0000000000000000 Success
//              A AMUT FFFFFFFFFFFFFFFF Error
func (s *Bravia) SetAudioMute(on bool) error {
	answer, err := s.exec(newMessageFromBool(Control, AMUT, on))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetAudioMute E AMUT ################ Retrieves the audio mute status
//              A AMUT 0000000000000000 Not Muted
//              A AMUT 0000000000000001 Muted
//              A AMUT FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetAudioMute() (bool, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, AMUT, ""))
	if err != nil {
		return false, err
	}
	return answer.Bool()
}

// SetInput C INPT 000000010000XXXX Changes the input to HDMI (1–9999)
//          C INPT 000000040000XXXX Changes the input to Component (1–9999)
//          C INPT 000000050000XXXX Changes the input to Screen Mirroring (1–9999)
//          A INPT NNNNNNNNNNNNNNNN Not Found
//          A INPT 0000000000000000 Success
//          A INPT FFFFFFFFFFFFFFFF Error
func (s *Bravia) SetInput(inputType, inputNum int) error {
	answer, err := s.exec(newMessageFromInts(Control, INPT, inputType, inputNum))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetInput E INPT ################ Get current input
//          A INPT 000000010000XXXX HDMI (1–9999)
//          A INPT 000000040000XXXX Component (1–9999)
//          A INPT 000000050000XXXX Screen Mirroring (1–9999)
func (s *Bravia) GetInput() (int, int, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, INPT, ""))
	if err != nil {
		return 0, 0, err
	}
	return answer.Ints()
}

// SetPictureMute C PMUT 0000000000000000 Disables the picture mute state
//                C PMUT 0000000000000001 Turns the screen black (picture mute)
//                A PMUT 0000000000000000 Success
//                A PMUT FFFFFFFFFFFFFFFF Error
func (s *Bravia) SetPictureMute(on bool) error {
	answer, err := s.exec(newMessageFromBool(Control, PMUT, on))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetPictureMute E PMUT ################ Checks if picture mute is enabled
//                A PMUT 0000000000000000 Disabled (Picture mute off)
//                A PMUT 0000000000000001 Enabled (Picture mute on)
//                A PMUT FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetPictureMute() (bool, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, PMUT, ""))
	if err != nil {
		return false, err
	}
	return answer.Bool()
}

// TogglePictureMute C TPMU ################ Toggles picture mode
//  A0000000000000000 Success
//  AFFFFFFFFFFFFFFFF Error
func (s *Bravia) TogglePictureMute() error {
	answer, err := s.exec(newMessageFromString(Control, TPMU, ""))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetBroadcastAddress E BADR eth0############ Retrieves the broadcast IPv4 address of the specified interface
//                     A BADR 192.168.0.14####  Broadcast address pad on the right with "#"
//                     A BADR FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetBroadcastAddress(adaptor string) (string, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, BADR, adaptor))
	if err != nil {
		return "", err
	}
	return answer.String()
}

// GetMacAddress E MADR eth0############ Retrieves the MAC address of the specified interface
//               A MADR XXXXXXXXXXXX#### MAC address pad on the right with "#"
//               A MADR FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetMacAddress(adaptor string) (string, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, MADR, adaptor))
	if err != nil {
		return "", err
	}
	return answer.String()
}

// SetSceneSetting C SCEN XXXXXXXXXXXXXXXX Changes the Scene Setting.
//                 A SCEN 0000000000000000 Success
//                 A SCEN NNNNNNNNNNNNNNNN Not available for the current input
//                 A SCEN FFFFFFFFFFFFFFFF Error
// The parameter strings are case-sensitive and pad on the right with "#". e.g.) auto24pSync#####
// auto, auto24pSync, general
func (s *Bravia) SetSceneSetting(sceneSetting string) error {
	answer, err := s.exec(newMessageFromString(Control, SCEN, sceneSetting))
	if err != nil {
		return err
	}
	return answer.Error()
}

// GetSceneSetting E SCEN ################ Retrieves the current Scene Setting
//                 A SCEN XXXXXXXXXXXXXXXX Success with Scene Setting value
//                 A SCEN NNNNNNNNNNNNNNNN Not available for the current input
//                 A SCEN FFFFFFFFFFFFFFFF Error
func (s *Bravia) GetSceneSetting() (string, error) {
	answer, err := s.exec(newMessageFromString(Enquiry, SCEN, ""))
	if err != nil {
		return "", err
	}
	return answer.String()
}
