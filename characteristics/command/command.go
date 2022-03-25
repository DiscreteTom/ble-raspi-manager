package command

import (
	"encoding/json"
	"github/DiscreteTom/ble-raspi-manager/internal/shell"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

type command struct {
	UUID   string
	Cmd    string
	Output string
}

func NewCharacteristicConfig(serviceUUID uuid.UUID) bluetooth.CharacteristicConfig {
	cmdCharUUID := uuid.NewSHA1(serviceUUID, []byte("cmd"))
	cmdCharBleUUID, _ := bluetooth.ParseUUID(cmdCharUUID.String())

	currentCmd := &command{}
	readOutputFrom := 0

	return bluetooth.CharacteristicConfig{
		UUID:  cmdCharBleUUID,
		Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
		ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
			txt, err := json.Marshal(currentCmd)
			if err != nil {
				return txt, err
			}
			if readOutputFrom < len(txt) {
				readOutputFrom += 256
				return txt[readOutputFrom-256:], nil
			} else {
				return []byte{}, nil
			}
		},
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			newCmd := &command{}
			json.Unmarshal(value, &newCmd)
			go func() {
				output, err := shell.RunCommand(newCmd.Cmd)
				if err != nil {
					output = "Error: " + err.Error()
				}
				currentCmd.Output = output
				currentCmd.UUID = newCmd.UUID
				currentCmd.Cmd = newCmd.Cmd
				readOutputFrom = 0
			}()
		},
	}

}
