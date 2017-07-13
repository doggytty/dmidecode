package dmidecode

import (
	"testing"
	"fmt"
)

func TestDmiDecode_QueryBIOS(t *testing.T) {
	d := Instance()
	d.SetPassword("xxxxxx")
	biosPtr, lanPtr, err := d.QueryBIOS()

	if err != nil {
		fmt.Println("failed")
		return
	} else {
		fmt.Println(biosPtr.Characteristics)
		fmt.Println(biosPtr.BIOSRevision)
		fmt.Println(lanPtr.InstallableLanguagesNumber)
	}
}

func TestDmiDecode_QuerySystem(t *testing.T) {
	d := Instance()
	d.SetPassword("xxxxxx")
	info, err := d.QuerySystem()
	if err != nil {
		fmt.Println("query system failed!")
	} else {
		fmt.Println(info.Manufacturer)
	}
}

func TestDmiDecode_QueryProcessor(t *testing.T) {
	d := Instance()
	d.SetPassword("xxxxxx")
	info, err := d.QueryProcessor()
	if err != nil {
		fmt.Println("query system failed!")
	} else {
		fmt.Println(info.CoreCount)
	}
}

func TestDmiDecode_QueryBaseBoard(t *testing.T) {
	d := Instance()
	d.SetPassword("xxxxxx")
	info, err := d.QueryBaseBoard()
	if err != nil {
		fmt.Println("query system failed!")
	} else {
		fmt.Println(info.Manufacturer)
	}
}

func TestDmiDecode_QueryCache(t *testing.T) {
	d := Instance()
	d.SetPassword("xxxxxx")
	info, err := d.QueryCache()
	if err != nil {
		fmt.Println("query system failed!")
	} else {
		fmt.Println(len(info))
		for _,cache := range info {
			fmt.Println(cache.SocketDesignation)
		}
	}
}

func TestDmiDecode_QueryMemory(t *testing.T) {
	d := Instance()
	d.SetPassword("xxxxxx")
	info, err := d.QueryMemory()
	if err != nil {fmt.Println(info.MaximumCapacity)
		fmt.Println("query system failed!")
	} else {
		fmt.Println(info.MaximumCapacity)
		for _, cache := range info.MemoryList {
			fmt.Println(cache.Size)
		}
	}
}

