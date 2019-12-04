package bravia

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// Control Used to control or change values on the Display
	Control = 'C'
	// Enquiry Used to control or change values on the Display
	Enquiry = 'E'
	// Answer Used to send a reply back to the client from the Display
	Answer = 'A'
	// Notify Used to send an event to the client from the Display
	Notify = 'N'
)

// Possible Messages
const (
	IRCC = "IRCC"
	TPOW = "TPOW"
	TPMU = "TPMU"
	BADR = "BADR"
	MADR = "MADR"
	SCEN = "SCEN"

	// These value
	POWR = "POWR" // bool
	PMUT = "PMUT" // bool
	AMUT = "AMUT" // bool
	INPT = "INPT" // int, int
	VOLU = "VOLU" // int

	// Not part of the offocial spec, But unlikely to be used in future updates
	CONNECTED = "tcp\x00" // bool
)

const (
	HDMI            = 1
	Component       = 4
	ScreenMirroring = 5
)

var InputName map[int]string = inputName()

func inputName() map[int]string {
	return map[int]string{
		1: "HDMI",
		4: "Component",
		5: "Screen Mirroring",
	}
}

type Message struct {
	CommandType byte
	Command     string
	Paramaters  string
}

func newMessage(cmd string) *Message {
	if len(cmd) != 24 || cmd[0] != '*' || cmd[1] != 'S' || cmd[23] != '\n' {
		fmt.Printf("cmd '%s'\n", cmd)
		return nil
	}

	return &Message{
		CommandType: cmd[2],
		Command:     cmd[3:7],
		Paramaters:  cmd[7:23],
	}
}

func newMessageFromString(MessageType byte, Message, Paramaters string) *Message {
	if len(Paramaters) < 16 { // pad to 16 bytes
		Paramaters += strings.Repeat("#", 16-len(Paramaters))
	}
	return newMessage(fmt.Sprintf("*S%c%s%s\n", MessageType, Message, Paramaters))
}

func newMessageFromInts(MessageType byte, Message string, first, second int) *Message {
	return newMessageFromString(MessageType, Message, fmt.Sprintf("%08d%08d", first, second))
}

func newMessageFromBool(MessageType byte, Message string, on bool) *Message {
	switch on {
	case true:
		return newMessageFromInts(MessageType, Message, 0, 1)
	case false:
		return newMessageFromInts(MessageType, Message, 0, 0)
	default:
		return nil
	}
}

func (s *Message) String() (string, error) {
	switch s.Paramaters {
	case "FFFFFFFFFFFFFFFF", "NNNNNNNNNNNNNNNN":
		return "", errors.New(s.Paramaters)
	}
	return strings.TrimRight(s.Paramaters, "#"), nil
}

func (s *Message) Ints() (int, int, error) {
	var first, second int
	n, err := fmt.Sscanf(s.Paramaters, "%08d%08d", &first, &second)
	if err != nil || n != 2 {
		return 0, 0, errors.New(s.Paramaters)
	}
	return first, second, err
}

func (s *Message) Int() (int, error) {
	first, second, err := s.Ints()
	switch {
	case err != nil:
		return 0, err
	case 0 != first:
		return 0, errors.New(s.Paramaters)
	default:
		return second, nil
	}
}

func (s *Message) Bool() (bool, error) {
	first, second, err := s.Ints()
	switch {
	case err != nil:
		return false, err
	case first == 0 && second == 0:
		return false, nil
	case first == 0 && second == 1:
		return true, nil
	default:
		return false, errors.New(s.Paramaters)
	}
}

func (s *Message) Error() error {
	if s.Paramaters != "0000000000000000" {
		return errors.New(s.Paramaters)
	}
	return nil
}
