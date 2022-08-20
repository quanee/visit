package main

import (
	"math/rand"
	"strconv"
)

func uuid() string {
	return strconv.Itoa(rand.Int())
}
