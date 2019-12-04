package bravia

//  Possible values for SendIrMessage() function
const (
	IrDisplay      = 5
	IrHome         = 6
	IrOptions      = 7
	IrReturn       = 8
	IrUp           = 9
	IrDown         = 10
	IrRight        = 11
	IrLeft         = 12
	IrConfirm      = 13
	IrRed          = 14
	IrGreen        = 15
	IrYellow       = 16
	IrBlue         = 17
	IrNum1         = 18
	IrNum2         = 19
	IrNum3         = 20
	IrNum4         = 21
	IrNum5         = 22
	IrNum6         = 23
	IrNum7         = 24
	IrNum8         = 25
	IrNum9         = 26
	IrNum0         = 27
	IrVolumeUp     = 30
	IrVolumeDown   = 31
	IrMute         = 32
	IrChannelUp    = 33
	IrChannelDown  = 34
	IrSubtitle     = 35
	IrDOT          = 38
	IrPictureOff   = 50
	IrWide         = 61
	IrJump         = 62
	IrSyncMenu     = 76
	IrForward      = 77
	IrPlay         = 78
	IrRewind       = 79
	IrPrev         = 80
	IrStop         = 81
	IrNext         = 82
	IrPause        = 84
	IrFlashPlus    = 86
	IrFlashMinus   = 87
	IrTVPower      = 98
	IrAudio        = 99
	IrInput        = 101
	IrSleep        = 104
	IrSleepTimer   = 105
	IrVideo2       = 208
	IrPictureMode  = 110
	IrDemoSurround = 121
	IrHDMI1        = 124
	IrHDMI2        = 125
	IrHDMI3        = 126
	IrHDMI4        = 127
	IrActionMenu   = 129
	IrHelp         = 130
)

// SendIrMessage   C IRCC XXXXXXXXXXXXXXXX Sends codes like IR commands of remote controller.  Please refer to IR Messages for the details.
//                 A IRCC 0000000000000000 Success
//                 A IRCC FFFFFFFFFFFFFFFF Error
func (s *Bravia) SendIrMessage(param int) error {
	answer, err := s.exec(newMessageFromInts(Control, IRCC, 0, param))
	if err != nil {
		return err
	}
	return answer.Error()
}
