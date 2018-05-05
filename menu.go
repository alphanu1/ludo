package main

import (
	"fmt"
	"io/ioutil"
	"libretro"
	"path/filepath"

	"github.com/fatih/structs"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
)

type menuCallback func()
type menuCallbackGetValue func() string

type entry struct {
	label         string
	value         string
	scroll        float32
	scrollTween   *gween.Tween
	ptr           int
	callback      menuCallback
	callbackValue menuCallbackGetValue
	callbackIncr  menuCallback
	children      []entry
}

var menuStack []entry

func buildExplorer(path string) entry {
	var menu entry
	menu.label = "Explorer"

	files, err := ioutil.ReadDir(path)
	if err != nil {
		notify(err.Error(), 240)
		fmt.Println(err)
	}

	for _, f := range files {
		f := f
		menu.children = append(menu.children, entry{
			label: f.Name(),
			callback: func() {
				if f.IsDir() {
					menuStack = append(menuStack, buildExplorer(path+"/"+f.Name()+"/"))
				} else if filepath.Ext(f.Name()) == ".dylib" {
					coreLoad(path + "/" + f.Name())
				} else if stringInSlice(filepath.Ext(f.Name()), []string{".sms", ".zip", ".sfc", ".md", ",bin", ".nes", ".pce"}) {
					coreLoadGame(path + "/" + f.Name())
				}
			},
		})
	}

	return menu
}

func buildSettings() entry {
	var menu entry
	menu.label = "Settings"

	fields := structs.Fields(&settings)
	for _, f := range fields {
		f := f
		menu.children = append(menu.children, entry{
			label: f.Tag("label"),
			value: fmt.Sprintf(f.Tag("fmt"), f.Value()),
			callbackIncr: func() {
				incrCallbacks[f.Name()](f)
			},
			callbackValue: func() string {
				return fmt.Sprintf(f.Tag("fmt"), f.Value())
			},
		})
	}

	return menu
}

func buildMainMenu() entry {
	var menu entry
	menu.label = "Main Menu"

	if g.coreRunning {
		menu.children = append(menu.children, entry{
			label: "Quick Menu",
			callback: func() {
				menuStack = append(menuStack, buildQuickMenu())
			},
		})
	}

	menu.children = append(menu.children, entry{
		label: "Load Core",
		callback: func() {
			menuStack = append(menuStack, buildExplorer("./cores"))
		},
	})

	menu.children = append(menu.children, entry{
		label: "Load Game",
		callback: func() {
			menuStack = append(menuStack, buildExplorer("./roms"))
		},
	})

	menu.children = append(menu.children, entry{
		label: "Settings",
		callback: func() {
			menuStack = append(menuStack, buildSettings())
		},
	})

	menu.children = append(menu.children, entry{
		label: "Help",
		callback: func() {
			notify("Not implemented yet", 240)
		},
	})

	menu.children = append(menu.children, entry{
		label: "Quit",
		callback: func() {
			window.SetShouldClose(true)
		},
	})

	return menu
}

func buildQuickMenu() entry {
	var menu entry
	menu.label = "Quick Menu"

	menu.children = append(menu.children, entry{
		label: "Resume",
		callback: func() {
			g.menuActive = !g.menuActive
		},
	})

	menu.children = append(menu.children, entry{
		label: "Reset",
		callback: func() {
			g.core.Reset()
			g.menuActive = false
		},
	})

	menu.children = append(menu.children, entry{
		label: "Save State",
		callback: func() {
			fmt.Println("[Menu]: Not implemented")
			notify("Not implemented", 240)
		},
	})

	menu.children = append(menu.children, entry{
		label: "Load State",
		callback: func() {
			fmt.Println("[Menu]: Not implemented")
			notify("Not implemented", 240)
		},
	})

	menu.children = append(menu.children, entry{
		label: "Take Screenshot",
		callback: func() {
			fmt.Println("[Menu]: Not implemented")
			notify("Not implemented", 240)
		},
	})

	return menu
}

var vSpacing = 70
var inputCooldown = 0

func menuInput() {
	currentMenu := &menuStack[len(menuStack)-1]

	if inputCooldown > 0 {
		inputCooldown--
	}

	if newState[0][libretro.DeviceIDJoypadDown] && inputCooldown == 0 {
		currentMenu.ptr++
		if currentMenu.ptr >= len(currentMenu.children) {
			currentMenu.ptr = 0
		}
		currentMenu.scrollTween = gween.New(currentMenu.scroll, float32(currentMenu.ptr*vSpacing), 0.15, ease.OutSine)
		inputCooldown = 10
	}

	if newState[0][libretro.DeviceIDJoypadUp] && inputCooldown == 0 {
		currentMenu.ptr--
		if currentMenu.ptr < 0 {
			currentMenu.ptr = len(currentMenu.children) - 1
		}
		currentMenu.scrollTween = gween.New(currentMenu.scroll, float32(currentMenu.ptr*vSpacing), 0.10, ease.OutSine)
		inputCooldown = 10
	}

	if released[0][libretro.DeviceIDJoypadA] {
		if currentMenu.children[currentMenu.ptr].callback != nil {
			currentMenu.children[currentMenu.ptr].callback()
		}
	}

	if released[0][libretro.DeviceIDJoypadRight] {
		if currentMenu.children[currentMenu.ptr].callbackIncr != nil {
			currentMenu.children[currentMenu.ptr].callbackIncr()
		}
	}

	if released[0][libretro.DeviceIDJoypadB] {
		if len(menuStack) > 1 {
			menuStack = menuStack[:len(menuStack)-1]
		}
	}
}

func renderMenuList() {
	w, h := window.GetFramebufferSize()
	fullscreenViewport()

	currentMenu := &menuStack[len(menuStack)-1]
	if currentMenu.scrollTween != nil {
		currentMenu.scroll, _ = currentMenu.scrollTween.Update(1.0 / 60.0)
	}

	video.font.SetColor(1, 1, 1, 1.0)
	video.font.Printf(60, 20+60, 0.5, currentMenu.label)

	for i, e := range currentMenu.children {
		y := -currentMenu.scroll + 20 + float32(vSpacing*(i+2))

		if y < 0 || y > float32(h) {
			continue
		}

		if i == currentMenu.ptr {
			video.font.SetColor(0.0, 1.0, 0.0, 1.0)
		} else {
			video.font.SetColor(0.6, 0.6, 0.9, 1.0)
		}
		video.font.Printf(100, y, 0.5, e.label)

		if e.callbackValue != nil {
			video.font.Printf(float32(w)-250, y, 0.5, e.callbackValue())
		}
	}
}

func menuInit() {
	if g.coreRunning {
		menuStack = append(menuStack, buildMainMenu())
		menuStack = append(menuStack, buildQuickMenu())
	} else {
		menuStack = append(menuStack, buildMainMenu())
	}
}
