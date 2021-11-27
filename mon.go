package main

import (
	"fmt"
	"os"
	"time"
)

var (
	ledGreenGPIO = 23
	ledRedGPIO   = 45
	greenOn      = false
	greenBlink   = false
	redOn        = false
	redBlink     = false
)

func setGPIO(gpio int, on bool) {

	// Open file
	fd, err := os.OpenFile(fmt.Sprintf("/sys/class/gpio/gpio%d/value", gpio), os.O_RDWR|os.O_SYNC, 0666)
	if err != nil {
		fmt.Printf("Unable to open GPIO %d. Skipping.", gpio)
		return
	}
	defer fd.Close()

	// Write value
	if on {
		_, err = fmt.Fprintln(fd, "1")
		if err != nil {
			fmt.Printf("Unable to write to GPIO %d. Skipping.", gpio)
			return
		}
	} else {
		_, err = fmt.Fprintln(fd, "0")
		if err != nil {
			fmt.Printf("Unable to write to GPIO %d. Skipping.", gpio)
			return
		}
	}
}

func SetGreenLed(on bool, blink bool) {
	greenOn = on
	greenBlink = blink
}

func SetRedLed(on bool, blink bool) {
	redOn = on
	redBlink = blink
}

func StartLed() {

	isNowGreenOn := false
	isNowRedOn := false

	go (func() {
		for {
			// Update state every second
			time.Sleep(time.Second)

			// Green
			if greenOn {
				if greenBlink {
					if isNowGreenOn {
						isNowGreenOn = false
						setGPIO(ledGreenGPIO, false)
					} else {
						isNowGreenOn = true
						setGPIO(ledGreenGPIO, true)
					}
				} else {
					isNowGreenOn = true
					setGPIO(ledGreenGPIO, true)
				}
			} else {
				isNowGreenOn = false
				setGPIO(ledGreenGPIO, false)
			}

			// Red
			if redOn {
				if redBlink {
					if isNowRedOn {
						isNowRedOn = false
						setGPIO(ledRedGPIO, false)
					} else {
						isNowRedOn = true
						setGPIO(ledRedGPIO, true)
					}
				} else {
					isNowRedOn = true
					setGPIO(ledRedGPIO, true)
				}
			} else {
				isNowRedOn = false
				setGPIO(ledRedGPIO, false)
			}
		}
	})()
}
