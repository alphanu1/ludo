package core

import (
	"log"
	"os"
	"os/user"
	"time"
	"unsafe"

	"github.com/libretro/ludo/libretro"
	"github.com/libretro/ludo/options"
	"github.com/libretro/ludo/settings"
	"github.com/libretro/ludo/state"
)

var logLevels = map[uint32]string{
	libretro.LogLevelDebug: "DEBUG",
	libretro.LogLevelInfo:  "INFO",
	libretro.LogLevelWarn:  "WARN",
	libretro.LogLevelError: "ERROR",
	libretro.LogLevelDummy: "DUMMY",
}

func logCallback(level uint32, str string) {
	log.Printf("[%s]: %s", logLevels[level], str)
}

func getTimeUsec() int64 {
	//fmt.Printf("Seconds since epoch %d", time.Now().Unix())
	return time.Now().UnixNano()
}

func environment(cmd uint32, data unsafe.Pointer) bool {
	switch cmd {
	case libretro.EnvironmentGetUsername:
		currentUser, err := user.Current()
		if err != nil {
			libretro.SetString(data, "")
		} else {
			libretro.SetString(data, currentUser.Username)
		}
	case libretro.EnvironmentGetLogInterface:
		state.Global.Core.BindLogCallback(data, logCallback)
	case libretro.EnvironmentGetPerfInterface:
		state.Global.Core.BindPerfCallback(data, getTimeUsec)
	case libretro.EnvironmentSetFrameTimeCallback:
		state.Global.Core.SetFrameTimeCallback(data)
	case libretro.EnvironmentSetAudioCallback:
		state.Global.Core.SetAudioCallback(data)
	case libretro.EnvironmentGetCanDupe:
		libretro.SetBool(data, true)
	case libretro.EnvironmentSetPixelFormat:
		format := libretro.GetPixelFormat(data)
		if format > libretro.PixelFormatRGB565 {
			return false
		}
		return vid.SetPixelFormat(format)
	case libretro.EnvironmentGetSystemDirectory:
		os.MkdirAll(settings.Current.SystemDirectory, os.ModePerm)
		libretro.SetString(data, settings.Current.SystemDirectory)
	case libretro.EnvironmentGetSaveDirectory:
		os.MkdirAll(settings.Current.SavefilesDirectory, os.ModePerm)
		libretro.SetString(data, settings.Current.SavefilesDirectory)
	case libretro.EnvironmentShutdown:
		vid.Window.SetShouldClose(true)
	case libretro.EnvironmentGetVariable:
		variable := libretro.GetVariable(data)
		for i, v := range Options.Vars {
			if variable.Key() == v.Key() {
				variable.SetValue(v.Choices()[Options.Choices[i]])
				return true
			}
		}
		return false
	case libretro.EnvironmentSetVariables:
		Options = options.New(libretro.GetVariables(data))
		return true
	case libretro.EnvironmentGetVariableUpdate:
		libretro.SetBool(data, Options.Updated)
		Options.Updated = false
		return true
	default:
		//log.Println("[Env]: Not implemented:", cmd)
		return false
	}
	return true
}
