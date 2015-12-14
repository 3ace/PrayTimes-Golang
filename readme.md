# PrayTimes (Golang)

This is a Golang port of PrayTimes.org's js code (ver 2.3)

This library can be used to calculate muslim's prayer time by time and location

### Exposed Functions

Use the following functions to interact with the library

```Go
    func GetTimes(date time.Time, coordinates []float64
            [, timeZone float64
            [, dst int
            [, timeFormat string]]]) map[string]string

    // set calculation method
	func SetMethod(method string)
	
	// adjust calculation parameters
	func Adjust(parameters map[string]interface{})
	
   // tune times by given offsets
	func Tune(offsets []int)

    // get calculation method
	func GetMethod() string
	
	// get current calculation parameters
	func GetSetting() map[string]interface{}
	
    // get current time offsets
	func GetOffsets() []int
```

### Use Example

```Go
// get today's prayer schedule for Bandung/Indonesia
pt := praytimes.GetTimes(time.Now(), []float64{ -6.9034443, 107.5731164 }, 7)
fmt.Printf("Sunrise = %v", times["sunrise"])
```


License
----

Permission is granted to use this code, with or
without modification, in any website or application
provided that credit is given to the original work
with a link back to PrayTimes.org.

This program is distributed in the hope that it will
be useful, but WITHOUT ANY WARRANTY.