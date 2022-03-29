package command

import (
	"github/DiscreteTom/ble-raspi-manager/internal/shell"
	"github/DiscreteTom/ble-raspi-manager/internal/transport"

	"github.com/google/uuid"
	"tinygo.org/x/bluetooth"
)

func NewCharacteristicConfig(serviceUUID uuid.UUID) bluetooth.CharacteristicConfig {
	cmdCharUUID := uuid.NewSHA1(serviceUUID, []byte("cmd"))
	cmdCharBleUUID, _ := bluetooth.ParseUUID(cmdCharUUID.String())

	var reader *transport.ReadHandler
	writer := transport.NewWriteHandler(func(uuid, content []byte) {
		go func() {
			output, err := shell.RunCommand(string(content))
			if err != nil {
				output = "Error: " + err.Error()
			}

			reader = transport.NewReadHandler(uuid, []byte(output))
		}()
	})

	return bluetooth.CharacteristicConfig{
		UUID:  cmdCharBleUUID,
		Flags: bluetooth.CharacteristicWritePermission | bluetooth.CharacteristicReadPermission,
		ReadEvent: func(client bluetooth.Connection) ([]byte, error) {
			return reader.Read(), nil
		},
		WriteEvent: func(client bluetooth.Connection, offset int, value []byte) {
			writer.Write(value)
		},
	}
}
