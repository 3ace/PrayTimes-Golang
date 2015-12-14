package praytimes

import (
    "testing"
    "fmt"
    "time"
    "strings"
)

func TestGetTimes(t *testing.T) {
    times := GetTimes(time.Now(), []float64 {43, -80}, -5, 0, "")

    fmt.Println("Prayer Times for today in Waterloo/Canada\n")
    for _, val := range []string {"Fajr", "Sunrise", "Dhuhr", "Asr", "Maghrib", "Isha", "Midnight"} {
        fmt.Printf("%v: %v\n", val, times[strings.ToLower(val)])
    }

    fmt.Println()
    times = GetTimes(time.Now(), []float64{ -6.9034443, 107.5731164 }, 7)
    fmt.Println("Prayer Times for today in Bandung/Indonesia\n")
    for _, val := range []string {"Fajr", "Sunrise", "Dhuhr", "Asr", "Maghrib", "Isha", "Midnight"} {
        fmt.Printf("%v: %v\n", val, times[strings.ToLower(val)])
    }

    fmt.Println()
}
