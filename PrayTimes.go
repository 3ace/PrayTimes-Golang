//--------------------- Copyright Block ----------------------
/*

PrayTimes.go: Prayer Times Calculator (ver 2.3)
Copyright (C) 2007-2015 PrayTimes.org

Go port by: Ade Anom
Original JS Code Developer: Hamid Zarrabi-Zadeh
License: GNU LGPL v3.0

TERMS OF USE:
	Permission is granted to use this code, with or
	without modification, in any website or application
	provided that credit is given to the original work
	with a link back to PrayTimes.org.

This program is distributed in the hope that it will
be useful, but WITHOUT ANY WARRANTY.

PLEASE DO NOT REMOVE THIS COPYRIGHT BLOCK.

*/


//--------------------- Help and Manual ----------------------
/*

User's Manual:
http://praytimes.org/manual

Calculation Formulas:
http://praytimes.org/calculation



//------------------------ User Interface -------------------------


	func GetTimes(date time.Time, coordinates []float64
            [, timeZone float64
            [, dst int
            [, timeFormat string]]]) map[string]string

	func SetMethod(method string)                          // set calculation method
	func Adjust(parameters map[string]interface{})         // adjust calculation parameters
	func Tune(offsets []int)                               // tune times by given offsets

	func GetMethod() string                                // get calculation method
	func GetSetting() map[string]interface{}               // get current calculation parameters
	func GetOffsets() []int                                // get current time offsets


//------------------------- Sample Usage --------------------------


	pt := praytimes.GetTimes(time.Now(), []float64{ -6.9034443, 107.5731164 }, 7, 0, "")
    fmt.Printf("Sunrise = %v", times["sunrise"])

*/


//----------------------- PrayTimes Class ------------------------

package praytimes

import (
    "fmt"
    "math"
    "strconv"
    "time"
    "strings"
    "regexp"
)

type CalculationMethod struct {
    name    string
    param   map[string]interface{}
}

var (
    // Calculation Methods
    methods map[string]CalculationMethod = map[string]CalculationMethod {
        "MWL" : CalculationMethod {
                "Muslim World League",
                map[string]interface{} {"fajr": 18, "isha":17},
            },
        "ISNA" : CalculationMethod {
                "Islamic Society of North America (ISNA)",
                map[string]interface{} {"fajr": 15, "isha": 15},
            },
        "Egypt" : CalculationMethod {
                "Egyptian General Authority of Survey",
                map[string]interface{} {"fajr": 19.5, "isha": 17.5},
            },
        "Makkah" : CalculationMethod {
                "Umm Al-Qura University, Makkah",
                map[string]interface{} {"fajr": 18.5, "isha": "90 min"},     // fajr was 19 degrees before 1430 hijri
            },
        "Karachi" : CalculationMethod {
                "University of Islamic Sciences, Karachi",
                map[string]interface{} {"fajr": 18, "isha": 18},
            },
        "Tehran" : CalculationMethod {
                "Institute of Geophysics, University of Tehran",
                map[string]interface{} {"fajr": 17.7, "isha": 14, "maghrib": 4.5, "midnight": "Jafari"},     // isha is not explicitly specified in this method
            },
        "Jafari" : CalculationMethod {
                "Shia Ithna-Ashari, Leva Institute, Qum",
                map[string]interface{} {"fajr": 16, "isha": 14, "maghrib": 4, "midnight": "Jafari"},
            },
    }

    timeNames []string = []string {
        "Imsak",
        "Fajr",
        "Sunrise",
        "Dhuhr",
        "Asr",
        "Sunset",
        "Maghrib",
        "Isha",
        "Midnight",
    }

    // Default Parameters in Calculation Methods
    defaultParams map[string]interface{} = map[string]interface{} {
        "maghrib": "0 min",
        "midnight": "Standard",
    }
)

//---------------------- Default Settings --------------------

var (
    calcMethod string = "MWL"

    // do not change anything here; use adjust method instead
    setting map[string]interface{} = map[string]interface{}  {
        "imsak"       : "10 min",
        "dhuhr"       : "0 min",
        "asr"         : "Standard",
        "highLats"    : "NightMiddle",
        "midnight"    : "Standard",
    }

    timeFormat string = "24h"
    timeSuffixes []string = []string{"am", "pm"}
    invalidTime string =  "-----"

    numIterations int = 1
    offset []int

//----------------------- Local Variables ---------------------

    // coordinates
    lat float64
    lng float64
    elv float64

    // time variables
    timeZone float64
    jDate float64
)

//---------------------- Initialization -----------------------

func init() {
    // set method defaults
    for _, method := range methods {
        for defkey, defval := range defaultParams {
            if _, ok := method.param[defkey]; !ok {
                method.param[defkey] = defval
            }
        }
    }

    // initialize settings
    params := methods[calcMethod].param
    for key, val := range params {
        setting[key] = val
    }

    offset = make([]int, 9)
}

//----------------------- Public Functions ------------------------

// set calculation method
func SetMethod(method string) {
    if m, ok := methods[method]; ok {
        Adjust(m.param)
        calcMethod = method
    }
}

// set calculating parameters
func Adjust(param map[string]interface{}) {
    setting = param
}

// set time offsets
func Tune(timeOffsets []int) {
    offset = timeOffsets
}

// get current calculation method
func GetMethod() string {
    return calcMethod
}

// get current setting
func GetSetting() map[string]interface{} {
    return setting
}

// get current time offsets
func GetOffsets() []int {
    return offset
}

// get default calc parametrs
func GetDefaults() map[string]CalculationMethod {
    return methods
}

// return prayer times for a given date
// timezone float64, dst int, format string
func GetTimes(date time.Time, coords []float64, args ...interface{}) map[string]string {
    lat = coords[0]
    lng = coords[1]

    timezone := 0.0
    dst := 0
    format := ""

    if len(args) > 0 {
        a, ok := args[0].(float64)

        if ok {
            timezone = a
        } else {
            b, ok := args[0].(int)

            if ok {
                timezone = float64(b)
            }
        }

        if len(args) > 1 {
            dst = args[1].(int)
        }

        if len(args) > 2 {
            format = args[2].(string)
        }
    }

    if(len(coords) > 2) {
        elv = coords[2]
    } else {
        elv = 0
    }

    if format != "" {
        timeFormat = format
    }

    timeZone = timezone
    if dst > 0 {
        timeZone++;
    }

    jDate = julian(date.Year(), int(date.Month()), date.Day()) - lng / (15 * 24)

    return computeTimes()
}

// convert float time to the given format (see timeFormats)
func GetFormattedTime(time float64, format string, suffixes []string) string {
    if math.IsNaN(time) {
        return invalidTime
    }

    if format == "Float" {
        return strconv.FormatFloat(time, 'G', -1, 64)
    }

    if len(suffixes) <= 0 {
        suffixes = timeSuffixes
    }

    time = fixHour(time + 0.5 / 60) // add 0.5 minutes to round

    hours := math.Floor(time)
    minutes := math.Floor((time - hours) * 60)
    suffix := ""
    formattedTime := ""

    if format == "12h" {
        if hours < 12 {
            suffix = suffixes[0]
        } else {
            suffix = suffixes[1]
        }
    }

    if format == "24h" {
        formattedTime = fmt.Sprintf("%02.f:%02.f", hours, minutes)
    } else {
        formattedTime = fmt.Sprintf("%02d:%02.f", ((int(hours) + 12 - 1) % 12 + 1), minutes)
    }

    return formattedTime + suffix
}

//---------------------- Calculation Functions -----------------------


// compute mid-day time
func midDay(time float64) float64 {
    _, eqt := sunPosition(jDate + time);
    return fixHour(12 - eqt)
}

// compute the time at which sun reaches a specific angle below horizon
func sunAngleTime(angle, time float64, direction ...string) float64 {
    decl, _ := sunPosition(jDate + time)
    noon := midDay(time)

    t := 1 / 15.0 * arccos((-sin(angle) - sin(decl) * sin(lat)) / (cos(decl) * cos(lat)))

    if len(direction) > 0 && direction[0] == "ccw" {
        return noon + -t
    }

    return noon + t
}

// compute asr time
func asrTime(factor float64, time float64) float64 {
    decl, _ := sunPosition(jDate + time)
    angle := -arccot(factor + tan(math.Abs(lat - decl)))

    return sunAngleTime(angle, time)
}

// compute declination angle of sun and equation of time
// Ref: http://aa.usno.navy.mil/faq/docs/SunApprox.php
func sunPosition(jd float64) (decl float64, eqt float64) {
    D := jd - 2451545.0
    g := fixAngle(357.529 + 0.98560028 * D)
    q := fixAngle(280.459 + 0.98564736 * D)
    L := fixAngle(q + 1.915 * sin(g) + 0.020 * sin(2*g));

    // R := 1.00014 - 0.01671 * cos(g) - 0.00014 * cos(2*g);
    e := 23.439 - 0.00000036 * D;

    RA := arctan2(cos(e) * sin(L), cos(L))/ 15;
    eqt = q/15 - fixHour(RA);
    decl = arcsin(sin(e) * sin(L));

    return decl, eqt;
}

// convert Gregorian date to Julian day
// Ref: Astronomical Algorithms by Jean Meeus
func julian(year, month, day int) float64 {
    if month <= 2 {
        year -= 1
        month += 12
    }

    A := math.Floor(float64(year) / 100)
    B := 2 - A + math.Floor(A / 4)

    return math.Floor(365.25 * (float64(year) + 4716)) + math.Floor(30.6001 * (float64(month) + 1)) + float64(day) + B - 1524.5
}


//---------------------- Compute Prayer Times -----------------------


// compute prayer times at given julian date
func computePrayerTimes(times map[string]float64) map[string]float64 {
    times = dayPortion(times)
    params := setting

    imsak := sunAngleTime(eval(params["imsak"]), times["imsak"], "ccw")
    fajr := sunAngleTime(eval(params["fajr"]), times["fajr"], "ccw")
    sunrise := sunAngleTime(riseSetAngle(), times["sunrise"], "ccw")
    dhuhr := midDay(times["dhuhr"])
    asr := asrTime(float64(asrFactor(params["asr"])), times["asr"])
    sunset := sunAngleTime(riseSetAngle(), times["sunset"])
    maghrib := sunAngleTime(eval(params["maghrib"]), times["maghrib"])
    isha := sunAngleTime(eval(params["isha"]), times["isha"])

    return map[string]float64 {
        "imsak" : imsak,
        "fajr" : fajr,
        "sunrise" : sunrise,
        "dhuhr" : dhuhr,
        "asr" : asr,
        "sunset" : sunset,
        "maghrib" : maghrib,
        "isha" : isha,
    }
}

// compute prayer times
func computeTimes() map[string]string {
    // default times
    times := map[string]float64 {
        "imsak" : 5,
        "fajr" : 5,
        "sunrise" : 6,
        "dhuhr" : 12,
        "asr" : 13,
        "sunset" : 18,
        "maghrib" : 18,
        "isha" : 18,
    }

    // main iterations
    for i := 1; i <= numIterations; i++ {
        times = computePrayerTimes(times)
    }

    times = adjustTimes(times)

    // add midnight time
    if setting["midnight"] == "Jafari" {
        times["midnight"] = times["sunset"] + timeDiff(times["sunset"], times["fajr"]) / 2
    } else {
        times["midnight"] = times["sunset"] + timeDiff(times["sunset"], times["sunrise"]) / 2
    }

    times = tuneTimes(times)
    return modifyFormats(times)
}

// adjust times
func adjustTimes(times map[string]float64) map[string]float64 {
    params := setting

    for key, _ := range times {
        times[key] += timeZone - lng / 15
    }

    if params["highLats"] != "None" {
        times = adjustHighLats(times)
    }

    if isMin(params["imsak"]) {
        times["imsak"] = times["fajr"] - eval(params["imsak"]) / 60
    }

    if isMin(params["maghrib"]) {
        times["maghrib"] = times["sunset"] + eval(params["maghrib"]) / 60
    }

    if isMin(params["isha"]) {
        times["isha"] = times["maghrib"] + eval(params["isha"]) / 60
    }

    times["dhuhr"] += eval(params["dhuhr"]) / 60

    return times
}

// get asr shadow factor
func asrFactor(asrParam interface{}) int {
    switch asrParam.(type) {
    case string:
        factor := map[string]int {
            "Standard" : 1,
            "Hanafi" : 2,
        }[asrParam.(string)]

        return factor
    }

    return int(eval(asrParam))
}

// return sun angle for sunset/sunrise
func riseSetAngle() float64 {
    angle := 0.0347 * math.Sqrt(elv); // an approximation
    return 0.833 + angle;
}

// apply offsets to the times
func tuneTimes(times map[string]float64) map[string]float64 {
    i := 0;
    for key, _ := range times {
        times[key] += float64(offset[i]) / 60
        i++
    }

    return times
}

// convert times to given time format
func modifyFormats(times map[string]float64) map[string]string {
    formatted := map[string]string{}

    for key, _ := range times {
        formatted[key] = GetFormattedTime(times[key], timeFormat, []string{})
    }

    return formatted
}

// adjust times for locations in higher latitudes
func adjustHighLats(times map[string]float64) map[string]float64 {
    params := setting
    nightTime := timeDiff(times["sunset"], times["sunrise"])

    times["imsak"] = adjustHLTime(times["imsak"], times["sunrise"], eval(params["imsak"]), nightTime, "ccw")
    times["fajr"]  = adjustHLTime(times["fajr"], times["sunrise"], eval(params["fajr"]), nightTime, "ccw");
    times["isha"]  = adjustHLTime(times["isha"], times["sunset"], eval(params["isha"]), nightTime);
    times["maghrib"] = adjustHLTime(times["maghrib"], times["sunset"], eval(params["maghrib"]), nightTime);

    return times
}

// adjust a time for higher latitudes
func adjustHLTime(time float64, base float64, angle float64,
    night float64, direction ...string) float64 {
    dir := ""

    if len(direction) > 0 {
        dir = direction[0]
    }

    portion := nightPortion(angle, night)
    td := 0.0

    if dir == "ccw" {
        td = timeDiff(time, base)
    } else {
        td = timeDiff(base, time)
    }

    if math.IsNaN(time) || td > portion {
        if dir == "ccw" {
            time = base - portion
        } else {
            time = base + portion
        }
    }

    return time
}

// the night portion used for adjusting times in higher latitudes
func nightPortion(angle float64, night float64) float64 {
    method := setting["highLats"];
    portion := 1.0 / 2.0                  //Midnight

    if method == "AngleBased" {
        portion = 1 / 60 * angle
    } else if method == "OneSeventh" {
        portion = 1/7
    }

    return portion * night
}

// convert hours to day portions
func dayPortion(times map[string]float64) map[string]float64 {
    for key, _ := range times {
        times[key] /= 24
    }

    return times
}

//---------------------- Time Zone Functions -----------------------


// get local time zone
func getTimeZone(date []int) int {
    year := date[0]
    t1 := float64(gmtOffset([]int{year, 0, 1}))
    t2 := float64(gmtOffset([]int{year, 6, 1}))

    return int(math.Min(t1, t2))
}

// get daylight saving for a given date
func getDst(date []int) bool {
    return gmtOffset(date) != getTimeZone(date)
}

// GMT offset for a given date
func gmtOffset(date []int) int {
    localDate := time.Date(date[0], time.Month(date[1] - 1), date[2], 12, 0, 0, 0, time.Local)
    // GMTString := localDate.UTC().String()
    GMTDate := localDate.UTC()
    hoursDiff := localDate.Sub(GMTDate).Hours()
    return int(hoursDiff)
}

//---------------------- Misc Functions -----------------------

// compute the difference between two times
func timeDiff(time1 float64, time2 float64) float64 {
    return fixHour(time2 - time1)
}

// convert given string into a number
func eval(st interface{}) float64 {
    switch st.(type) {
    case int:
        return float64(st.(int))
    case float64:
        return st.(float64)
    case string:
        re := regexp.MustCompile(`\d+`)
        s := re.FindString(st.(string))

        if f, err := strconv.ParseFloat(s, 64); err == nil {
            return f
        }
    }

    return 0.0
}

// detect if input contains "min"
func isMin(arg interface{}) bool {
    switch arg.(type) {
    case string:
        return strings.Contains(arg.(string), "min")

    default:
        return false
    }
}

//---------------------- Degree-Based Math Class -----------------------

func dtr(d float64) float64 {
    return (d * math.Pi) / 180.0
}

func rtd(r float64) float64 {
    return (r * 180.0) / math.Pi
}

func sin(d float64) float64 {
    return math.Sin(dtr(d))
}

func cos(d float64) float64 {
    return math.Cos(dtr(d))
}

func tan(d float64) float64 {
    return math.Tan(dtr(d))
}

func arcsin(d float64) float64 {
    return rtd(math.Asin(d))
}

func arccos(d float64) float64 {
    return rtd(math.Acos(d))
}

func arctan(d float64) float64 {
    return rtd(math.Atan(d))
}

func arccot(x float64) float64 {
    return rtd(math.Atan(1/x))
}

func arctan2(y, x float64) float64 {
    return rtd(math.Atan2(y, x))
}

func fixAngle(a float64) float64 {
    return fix(a, 360)
}

func fixHour(a float64) float64 {
    return fix(a, 24)
}

func fix(a, b float64) float64 {
    a = a - b * (math.Floor(a / b))

    if a < 0 {
        return a + b
    }

    return a
}
