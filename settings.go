package gopeer

type SettingsType map[string]interface{}
type settingsStruct struct {
	TITLE_CONNECT      string
	TITLE_DISCONNECT   string
	TITLE_FILETRANSFER string
	OPTION_GET         string
	OPTION_SET         string
	SERVER_NAME        string
	IS_CLIENT          string
	END_BYTES          string
	TEMPLATE           string
	HMACKEY            string
	NETWORK            string
	VERSION            string
	MAX_ID             uint64
	PACK_SIZE          uint32
	BUFF_SIZE          uint32
	REMEMBER           uint16
	DIFFICULTY         uint8
	WAITING_TIME       uint8
	REDIRECT_QUAN      uint8
}

var settings = defaultSettings()

func defaultSettings() settingsStruct {
	return settingsStruct{
		TITLE_CONNECT:      "[TITLE-CONNECT]",
		TITLE_DISCONNECT:   "[TITLE-DISCONNECT]",
		TITLE_FILETRANSFER: "[TITLE-FILETRANSFER]",
		OPTION_GET:         "[OPTION-GET]", // Send
		OPTION_SET:         "[OPTION-SET]", // Receive
		IS_CLIENT:          "[IS-CLIENT]",
		SERVER_NAME:        "GOPEER-FRAMEWORK",
		END_BYTES:          "\000\000\000\005\007\001\000\000\000",
		TEMPLATE:           "0.0.0.0",
		HMACKEY:            "PASSWORD",
		NETWORK:            "NETWORK-NAME",
		VERSION:            "Version 1.0.0",
		MAX_ID:             5,       // 2^32 packages
		PACK_SIZE:          8 << 20, // 8MiB
		BUFF_SIZE:          1 << 20, // 1MiB
		REMEMBER:           256,     // hash packages
		DIFFICULTY:         15,
		WAITING_TIME:       5, // seconds
		REDIRECT_QUAN:      3,
	}
}

func Set(settings SettingsType) []uint8 {
	var (
		list = make([]uint8, len(settings))
		i    = 0
	)

	for name, data := range settings {
		switch data.(type) {
		case string:
			list[i] = stringSettings(name, data)
		case int:
			list[i] = intSettings(name, data)
		default:
			list[i] = 2
		}
		i++
	}

	return list
}

func Get(key string) interface{} {
	switch key {
	case "TITLE_CONNECT":
		return settings.TITLE_CONNECT
	case "TITLE_DISCONNECT":
		return settings.TITLE_DISCONNECT
	case "TITLE_FILETRANSFER":
		return settings.TITLE_FILETRANSFER
	case "OPTION_GET":
		return settings.OPTION_GET
	case "OPTION_SET":
		return settings.OPTION_SET
	case "SERVER_NAME":
		return settings.SERVER_NAME
	case "IS_CLIENT":
		return settings.IS_CLIENT
	case "END_BYTES":
		return settings.END_BYTES
	case "NETWORK":
		return settings.NETWORK
	case "VERSION":
		return settings.VERSION
	case "TEMPLATE":
		return settings.TEMPLATE
	case "HMACKEY":
		return settings.HMACKEY
	case "MAX_ID":
		return settings.MAX_ID
	case "PACK_SIZE":
		return settings.PACK_SIZE
	case "BUFF_SIZE":
		return settings.BUFF_SIZE
	case "REMEMBER":
		return settings.REMEMBER
	case "DIFFICULTY":
		return settings.DIFFICULTY
	case "WAITING_TIME":
		return settings.WAITING_TIME
	case "REDIRECT_QUAN":
		return settings.REDIRECT_QUAN
	default:
		return nil
	}
}

func stringSettings(name string, data interface{}) uint8 {
	result := data.(string)
	switch name {
	case "TITLE_CONNECT":
		settings.TITLE_CONNECT = result
	case "TITLE_DISCONNECT":
		settings.TITLE_DISCONNECT = result
	case "TITLE_FILETRANSFER":
		settings.TITLE_FILETRANSFER = result
	case "OPTION_GET":
		settings.OPTION_GET = result
	case "OPTION_SET":
		settings.OPTION_SET = result
	case "SERVER_NAME":
		settings.SERVER_NAME = result
	case "IS_CLIENT":
		settings.IS_CLIENT = result
	case "END_BYTES":
		settings.END_BYTES = result
	case "NETWORK":
		settings.NETWORK = result
	case "VERSION":
		settings.VERSION = result
	case "TEMPLATE":
		settings.TEMPLATE = result
	case "HMACKEY":
		settings.HMACKEY = result
	default:
		return 1
	}
	return 0
}

func intSettings(name string, data interface{}) uint8 {
	result := data.(int)
	switch name {
	case "MAX_ID":
		settings.MAX_ID = uint64(result)
	case "PACK_SIZE":
		settings.PACK_SIZE = uint32(result)
	case "BUFF_SIZE":
		settings.BUFF_SIZE = uint32(result)
	case "REMEMBER":
		settings.REMEMBER = uint16(result)
	case "DIFFICULTY":
		settings.DIFFICULTY = uint8(result)
	case "WAITING_TIME":
		settings.WAITING_TIME = uint8(result)
	case "REDIRECT_QUAN":
		settings.REDIRECT_QUAN = uint8(result)
	default:
		return 1
	}
	return 0
}
