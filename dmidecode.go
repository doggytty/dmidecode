package dmidecode

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"regexp"
	"bytes"
)

// type: bios system baseboard chassis processor memory Cache connector slot

const DEBUG = true

// linux判断命令是否安装
func CommandExist(command string) (string, error) {
	path, err := exec.LookPath(command)
	if err != nil {
		if DEBUG {
			log.Println("命令未安装!")
		}
		return "", err
	}
	if DEBUG {
		log.Printf("命令已安装,路径:%s\n", path)
	}
	return path, nil
}

func ExecuteCommand(cmd string) (string, error) {
	tty := exec.Command("/bin/sh", "-c", cmd)
	// stdout
	var out bytes.Buffer
	tty.Stdout = &out
	// stderr
	var stdErr bytes.Buffer
	tty.Stderr = &stdErr
	// run command
	err := tty.Run()
	if err != nil {
		if DEBUG {
			log.Printf("error: %q\n", stdErr.String())
		}
		return "", err
	}
	if DEBUG {
		log.Printf("out: %q\n", out.String())
	}
	return out.String(), nil
}

type DmiDecode struct {
	Path     string // dmidecode的安装路径
	LastErr  error  // 错误信息
	IsRoot   bool   // 是否root用户启动
	IsActive bool   // 是否可用
	Password string // 使用sudo执行命令需要传入的password
}

func Instance() *DmiDecode {
	// 初始化
	d := new(DmiDecode)
	// 查找dmidecode的路径
	d.Path, d.LastErr = CommandExist("dmidecode")
	if d.LastErr != nil {
		d.Path = ""
	}
	if DEBUG {
		log.Printf("search dmidecode over. path: %s\n", d.Path)
	}
	// 判断当前用户是否root
	if os.Getuid() == 0 {
		d.IsRoot = true
	}
	// 判断当前模块是否可用
	if d.Path != "" && d.IsRoot {
		d.IsActive = true
	}
	return d
}

func (d *DmiDecode) SetPassword(password string) {
	d.Password = password
	// 设置当前模块是否可用
	if d.Path != "" {
		d.IsActive = true
	}
	// 重新设置d.path
	d.Path = fmt.Sprintf("echo %s | sudo -S %s ", password, d.Path)
	if DEBUG {
		log.Printf("renew path: %s\n", d.Path)
	}
}

// dmidecode -t bios
type BiosInfo struct {
	//Vendor: LENOVO
	Vendor          string
	//Version: J4ET76WW(1.76)
	Version         string
	//Address: 0xE0000
	Address         string
	//BIOS Revision: 1.76
	BIOSRevision    string
	//Release Date: 03/03/2015
	ReleaseDate     string
	//Runtime Size: 128 kB
	RuntimeSize     string
	//ROM Size: 8192 kB
	RomSize         string
	//Characteristics:
	Characteristics []string
}

type BiosLanguage struct {
	//Language Description Format: Abbreviated
	LanguageDescriptionFormat string
	//Installable Languages: 7
	InstallableLanguagesNumber int
	InstallableLanguages []string
	//Currently Installed Language: en-US
	CurrentlyInstalledLanguage string
}

func (d *DmiDecode) QueryBIOS() (*BiosInfo, *BiosLanguage, error) {
	cmd := fmt.Sprintf("%s -t bios", d.Path)
	if DEBUG {
		log.Println("now query bios info: " + cmd)
	}
	bios, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, nil, err
	}
	var result *BiosInfo = new(BiosInfo)
	var language *BiosLanguage = new(BiosLanguage)
	biosArray := strings.Split(bios, "\n\n")
	for _, biosInfo := range biosArray {
		if strings.Contains(biosInfo, "\nBIOS Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			biosInfo = re.ReplaceAllString(biosInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			biosInfoArray := re.FindAllString(biosInfo, -1)
			for _, subBiosInfo := range biosInfoArray {
				subBiosInfoArray := strings.Split(subBiosInfo, ":")
				if len(subBiosInfoArray) == 2 {
					key := strings.TrimSpace(subBiosInfoArray[0])
					value := strings.TrimSpace(subBiosInfoArray[1])
					switch key {
					case "Vendor":
						result.Vendor = value
					case "Version":
						result.Version = value
					case "Release Date":
						result.ReleaseDate = value
					case "Address":
						result.Address = value
					case "Runtime Size":
						result.RuntimeSize = value
					case "ROM Size":
						result.RomSize = value
					case "Characteristics":
						characterArray := strings.Split(value, "|")
						for index, subValue := range characterArray {
							characterArray[index] = strings.TrimSpace(subValue)
						}
						result.Characteristics = characterArray
					case "BIOS Revision":
						result.BIOSRevision = value
					}
				}
			}
		} else if strings.Contains(biosInfo, "\nBIOS Language Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			biosInfo = re.ReplaceAllString(biosInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			languageArray := re.FindAllString(biosInfo, -1)
			for _, subLanguage := range languageArray {
				subLanguageArray := strings.Split(subLanguage, ":")
				if len(subLanguageArray) == 2 {
					key := strings.TrimSpace(subLanguageArray[0])
					value := strings.TrimSpace(subLanguageArray[1])
					switch key {
					case "Language Description Format":
						language.LanguageDescriptionFormat = value
					case "Currently Installed Language":
						language.CurrentlyInstalledLanguage = value
					case "Installable Languages":
						tmpArray := strings.Split(value, "|")
						for index, subValue := range tmpArray {
							tmpArray[index] = strings.TrimSpace(subValue)
						}
						language.InstallableLanguagesNumber = len(tmpArray) - 1
						language.InstallableLanguages = tmpArray[1:]
					}
				}
			}
		}
	}
	return result, language, nil
}

// dmidecode -t system
type SystemInfo struct {
	Manufacturer string
	ProductName  string
	Version      string
	SerialNumber string
	UUID         string
	WakeUpType   string
	SKUNumber    string
	Family       string
}

func (d *DmiDecode) QuerySystem() (*SystemInfo, error) {
	cmd := fmt.Sprintf("%s -t System", d.Path)
	if DEBUG {
		log.Println("now query System info: " + cmd)
	}
	system, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result *SystemInfo = new(SystemInfo)
	systemArray := strings.Split(system, "\n\n")
	for _, systemInfo := range systemArray {
		if strings.Contains(systemInfo, "\nSystem Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			systemInfo = re.ReplaceAllString(systemInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			systemInfoArray := re.FindAllString(systemInfo, -1)
			for _, subSystemInfo := range systemInfoArray {
				subSystemInfoArray := strings.Split(subSystemInfo, ":")
				if len(subSystemInfoArray) == 2 {
					key := strings.TrimSpace(subSystemInfoArray[0])
					value := strings.TrimSpace(subSystemInfoArray[1])
					switch key {
					case "Manufacturer":
						result.Manufacturer = value
					case "Product Name":
						result.ProductName = value
					case "Version":
						result.Version = value
					case "Serial Number":
						result.SerialNumber = value
					case "UUID":
						result.UUID = value
					case "Wake-up Type":
						result.WakeUpType = value
					case "SKU Number":
						result.SKUNumber = value
					case "Family":
						result.Family = value
					}
				}
			}
		}
	}
	return result, nil
}

// dmidecode -t baseboard
type BaseBoardInfo struct {
	//Manufacturer: LENOVO
	Manufacturer string
	//Product Name: 20ASEB3
	ProductName string
	//Version: No DPK
	Version string
	//Serial Number: ZZ0R958AGF4
	SerialNumber string
	//Asset Tag: Not Available
	AssertTag string
	//Features:
	//Board is a hosting board
	//Board is replaceable
	Features []string
	//Location In Chassis: Not Available
	LocationInChassis string
	//Chassis Handle: 0x0000
	ChassisHandle string
	//Type: Motherboard
	Type string
	//Contained Object Handles: 0
	ContainedObjectHandles string
}

func (d *DmiDecode) QueryBaseBoard() (*BaseBoardInfo, error) {
	cmd := fmt.Sprintf("%s -t baseboard", d.Path)
	if DEBUG {
		log.Println("now query baseboard info: " + cmd)
	}
	baseBoard, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result *BaseBoardInfo = new(BaseBoardInfo)
	baseBoardArray := strings.Split(baseBoard, "\n\n")
	for _, baseBoardInfo := range baseBoardArray {
		if strings.Contains(baseBoardInfo, "\nBase Board Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			baseBoardInfo = re.ReplaceAllString(baseBoardInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			baseBoardInfoArray := re.FindAllString(baseBoardInfo, -1)
			for _, subBaseBoardInfo := range baseBoardInfoArray {
				subBaseBoardInfoArray := strings.Split(subBaseBoardInfo, ":")
				if len(subBaseBoardInfoArray) == 2 {
					key := strings.TrimSpace(subBaseBoardInfoArray[0])
					value := strings.TrimSpace(subBaseBoardInfoArray[1])
					switch key {
					case "Manufacturer":
						result.Manufacturer = value
					case "Product Name":
						result.ProductName = value
					case "Version":
						result.Version = value
					case "Serial Number":
						result.SerialNumber = value
					case "Asset Tag":
						result.AssertTag = value
					case "Location In Chassis":
						result.LocationInChassis = value
					case "Chassis Handle":
						result.ChassisHandle = value
					case "Type":
						result.Type = value
					case "Contained Object Handles":
						result.ContainedObjectHandles = value
					case "Features":
						featuresArray := strings.Split(value, "|")
						for index, subValue := range featuresArray {
							featuresArray[index] = strings.TrimSpace(subValue)
						}
						result.Features = featuresArray
					}
				}
			}
		}
	}
	return result, nil
}

// dmidecode -t chassis
type ChassisInfo struct {
	//Manufacturer: LENOVO
	Manufacturer string
	//Type: Notebook
	Type string
	//Lock: Not Present
	Lock string
	//Version: Not Available
	Version string
	//Serial Number: ZZ0R958AGF4
	SerialNumber string
	//Asset Tag: Not Available
	AssertTag string
	//Boot-up State: Unknown
	BootUpState string
	//Power Supply State: Unknown
	PowerSupplyState string
	//Thermal State: Unknown
	ThermalState string
	//Security Status: Unknown
	SecurityStatus string
	//OEM Information: 0x00000000
	OEMInformation string
	//Height: Unspecified
	Height string
	//Number Of Power Cords: Unspecified
	NumberOfPowerCords string
	//Contained Elements: 0
	ContainedElements string
	//SKU Number: Not Specified
	SKUNumber string
}

func (d *DmiDecode) QueryChassis() (*ChassisInfo, error) {
	cmd := fmt.Sprintf("%s -t chassis", d.Path)
	if DEBUG {
		log.Println("now query chassis info: " + cmd)
	}
	chassis, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result *ChassisInfo = new(ChassisInfo)
	chassisArray := strings.Split(chassis, "\n\n")
	for _, chassisInfo := range chassisArray {
		if strings.Contains(chassisInfo, "\nChassis Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			chassisInfo = re.ReplaceAllString(chassisInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			chassisInfoArray := re.FindAllString(chassisInfo, -1)
			for _, subChassisInfo := range chassisInfoArray {
				subChassisInfoArray := strings.Split(subChassisInfo, ":")
				if len(subChassisInfoArray) == 2 {
					key := strings.TrimSpace(subChassisInfoArray[0])
					value := strings.TrimSpace(subChassisInfoArray[1])
					switch key {
					case "Manufacturer":
						result.Manufacturer = value
					case "Type":
						result.Type = value
					case "Lock":
						result.Lock = value
					case "Version":
						result.Version = value
					case "Serial Number":
						result.SerialNumber = value
					case "Asset Tag":
						result.AssertTag = value
					case "Boot-up State":
						result.BootUpState = value
					case "Thermal State":
						result.ThermalState = value
					case "Power Supply State":
						result.PowerSupplyState = value
					case "Security Status":
						result.SecurityStatus = value
					case "OEM Information":
						result.SecurityStatus = value
					case "Height":
						result.SecurityStatus = value
					case "Number Of Power Cords":
						result.SecurityStatus = value
					case "Contained Elements":
						result.SecurityStatus = value
					case "SKU Number":
						result.SecurityStatus = value
					}
				}
			}
		}
	}
	return result, nil
}

// dmidecode -t processor
type ProcessorInfo struct {
	//Socket Designation: CPU Socket - U3E1
	SocketDesignation string
	//Type: Central Processor
	Type string
	//Family: Core i7
	Family string
	//Manufacturer: Intel(R) Corporation
	Manufacturer string
	//ID: C3 06 03 00 FF FB EB BF
	ID string
	//Signature: Type 0, Family 6, Model 60, Stepping 3
	Signature string
	//Flags
	Flags []string
	//Version: Intel(R) Core(TM) i7-4712MQ CPU @ 2.30GHz
	Version string
	//Voltage: 0.7 V
	Voltage string
	//External Clock: 100 MHz
	ExternalClock string
	//Max Speed: 2300 MHz
	MaxSpeed string
	//Current Speed: 2300 MHz
	CurrentSpeed string
	//Status: Populated, Enabled
	Status string
	//Upgrade: Socket rPGA988B
	Upgrade string
	//L1 Cache Handle: 0x0002
	L1CacheHandle string
	L2CacheHandle string
	L3CacheHandle string
	//Serial Number: To Be Filled By O.E.M.
	SerialNumber string
	//Asset Tag: To Be Filled By O.E.M.
	AssetTag string
	//Part Number: To Be Filled By O.E.M.
	PartNumber string
	//Core Count: 4
	CoreCount string
	//Core Enabled: 4
	CoreEnabled string
	//Thread Count: 8
	ThreadCount string
	//Characteristics:
	Characteristics []string
}

func (d *DmiDecode) QueryProcessor() (*ProcessorInfo, error) {
	cmd := fmt.Sprintf("%s -t processor", d.Path)
	if DEBUG {
		log.Println("now query chassis info: " + cmd)
	}
	processor, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result *ProcessorInfo = new(ProcessorInfo)
	processorArray := strings.Split(processor, "\n\n")
	for _, processorInfo := range processorArray {
		if strings.Contains(processorInfo, "\nProcessor Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			processorInfo = re.ReplaceAllString(processorInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			processorInfoArray := re.FindAllString(processorInfo, -1)
			for _, subProcessorInfo := range processorInfoArray {
				subProcessorInfoArray := strings.Split(subProcessorInfo, ":")
				if len(subProcessorInfoArray) == 2 {
					key := strings.TrimSpace(subProcessorInfoArray[0])
					value := strings.TrimSpace(subProcessorInfoArray[1])
					switch key {
					case "Socket Designation":
						result.SocketDesignation = value
					case "Type":
						result.Type = value
					case "Family":
						result.Family = value
					case "Manufacturer":
						result.Manufacturer = value
					case "ID":
						result.ID = value
					case "Signature":
						result.Signature = value
					case "Flags":
						flagsArray := strings.Split(value, "|")
						for index, subValue := range flagsArray {
							flagsArray[index] = strings.TrimSpace(subValue)
						}
						result.Flags = flagsArray
					case "Version":
						result.Version = value
					case "Voltage":
						result.Voltage = value
					case "External Clock":
						result.ExternalClock = value
					case "Max Speed":
						result.MaxSpeed = value
					case "Current Speed":
						result.CurrentSpeed = value
					case "Status":
						result.Status = value
					case "Upgrade":
						result.Upgrade = value
					case "L1 Cache Handle":
						result.L1CacheHandle = value
					case "L2 Cache Handle":
						result.L2CacheHandle = value
					case "Serial Number":
						result.SerialNumber = value
					case "Asset Tag":
						result.AssetTag = value
					case "Part Number":
						result.PartNumber = value
					case "Core Count":
						result.CoreCount = value
					case "Core Enabled":
						result.CoreEnabled = value
					case "Thread Count":
						result.ThreadCount = value
					case "Characteristics":
						characterArray := strings.Split(value, "|")
						for index, subValue := range characterArray {
							characterArray[index] = strings.TrimSpace(subValue)
						}
						result.Flags = characterArray
					}
				}
			}
		}
	}
	return result, nil
}

// dmidecode -t memory
type MemoryInfo struct {
	//Location: System Board Or Motherboard
	Location string
	//Use: System Memory
	Use string
	//Error Correction Type: None
	ErrorCorrectionType string
	//Maximum Capacity: 16 GB
	MaximumCapacity string
	//Error Information Handle: Not Provided
	ErrorInformationHandle string
	//Number Of Devices: 2
	NumberOfDevices string
	MemoryList []*MemoryDevice
}

type MemoryDevice struct {
	//Array Handle: 0x0005
	ArrayHandle string
	//Error Information Handle: Not Provided
	ErrorInformationHandle string
	//Total Width: 64 bits
	TotalWidth string
	//Data Width: 64 bits
	DataWidth string
	// Size: 4096 MB
	Size string
	//Form Factor: SODIMM
	FormFactor string
	//Set: None
	Set string
	//Locator: ChannelB-DIMM0
	Locator string
	//Bank Locator: BANK 2
	BankLocator string
	//Type: DDR3
	Type string
	//Type Detail: Synchronous
	TypeDetail string
	//Speed: 1600 MHz
	Speed string
	//Manufacturer: Hynix/Hyundai
	Manufacturer string
	//Serial Number: 1A6266B1
	SerialNumber string
	//Asset Tag: 9876543210
	AssetTag string
	//Part Number: HMT451S6AFR8A-PB
	PartNumber string
	//Rank: Unknown
	Rank string
	//Configured Clock Speed: 1600 MHz
	ConfiguredClockSpeed string
}

func (d *DmiDecode) QueryMemory() (*MemoryInfo, error) {
	cmd := fmt.Sprintf("%s -t memory", d.Path)
	if DEBUG {
		log.Println("now query memory info: " + cmd)
	}
	memory, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result *MemoryInfo = new(MemoryInfo)
	result.MemoryList = make([]*MemoryDevice, 0)
	memoryArray := strings.Split(memory, "\n\n")
	for _, memoryInfo := range memoryArray {
		if strings.Contains(memoryInfo, "\nPhysical Memory Array\n") {
			re, _ := regexp.Compile("\n\t\t")
			memoryInfo = re.ReplaceAllString(memoryInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			memoryInfoArray := re.FindAllString(memoryInfo, -1)
			for _, subMemoryInfo := range memoryInfoArray {
				subMemoryInfoArray := strings.Split(subMemoryInfo, ":")
				if len(subMemoryInfoArray) == 2 {
					key := strings.TrimSpace(subMemoryInfoArray[0])
					value := strings.TrimSpace(subMemoryInfoArray[1])
					switch key {
					case "Location":
						result.Location = value
					case "Use":
						result.Use = value
					case "Error Correction Type":
						result.ErrorCorrectionType = value
					case "Maximum Capacity":
						result.MaximumCapacity = value
					case "Error Information Handle":
						result.ErrorInformationHandle = value
					case "Number Of Devices":
						result.NumberOfDevices = value
					}
				}
			}
		} else if strings.Contains(memoryInfo, "\nMemory Device\n") {
			var memDevice *MemoryDevice = new(MemoryDevice)
			re, _ := regexp.Compile("\n\t\t")
			memoryInfo = re.ReplaceAllString(memoryInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			memoryDeviceArray := re.FindAllString(memoryInfo, -1)
			for _, subMemoryDevice := range memoryDeviceArray {
				subMemoryDeviceArray := strings.Split(subMemoryDevice, ":")
				if len(subMemoryDeviceArray) == 2 {
					key := strings.TrimSpace(subMemoryDeviceArray[0])
					value := strings.TrimSpace(subMemoryDeviceArray[1])
					switch key {
					case "Array Handle":
						memDevice.ArrayHandle = value
					case "Error Information Handle":
						memDevice.ErrorInformationHandle = value
					case "Total Width":
						memDevice.TotalWidth = value
					case "Data Width":
						memDevice.DataWidth = value
					case "Size":
						memDevice.Size = value
					case "Form Factor":
						memDevice.FormFactor = value
					case "Set":
						memDevice.Set = value
					case "Locator":
						memDevice.Locator = value
					case "Bank Locator":
						memDevice.BankLocator = value
					case "Type":
						memDevice.Type = value
					case "Type Detail":
						memDevice.TypeDetail = value
					case "Speed":
						memDevice.Speed = value
					case "Manufacturer":
						memDevice.Manufacturer = value
					case "Serial Number":
						memDevice.SerialNumber = value
					case "Asset Tag":
						memDevice.AssetTag = value
					case "Part Number":
						memDevice.PartNumber = value
					case "Rank":
						memDevice.Rank = value
					case "Configured Clock Speed":
						memDevice.ConfiguredClockSpeed = value
					}
				}
			}
			result.MemoryList = append(result.MemoryList, memDevice)
		}

	}
	return result, nil
}

// dmidecode -t cache
type CacheInfo struct {
	//Socket Designation: L2-Cache
	SocketDesignation string
	//Configuration: Enabled, Not Socketed, Level 2
	Configuration string
	//Operational Mode: Write Back
	OperationalMode string
	//Location: Internal
	Location string
	//Installed Size: 256 kB
	InstalledSize string
	//Maximum Size: 256 kB
	MaximumSize string
	//Supported SRAM Types:
	SupportedSRAMTypes []string
	//Asynchronous
	//Installed SRAM Type: Asynchronous
	InstalledSRAMType string
	//Speed: Unknown
	Speed string
	//Error Correction Type: Single-bit ECC
	ErrorCorrectionType string
	//System Type: Unified
	SystemType string
	//Associativity: 8-way Set-associative
	Associativity string
}

func (d *DmiDecode) QueryCache() ([]*CacheInfo, error) {
	cmd := fmt.Sprintf("%s -t cache", d.Path)
	if DEBUG {
		log.Println("now query cache info: " + cmd)
	}
	cache, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result = make([]*CacheInfo, 0)
	cacheArray := strings.Split(cache, "\n\n")
	for _, cacheInfo := range cacheArray {
		if strings.Contains(cacheInfo, "\nCache Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			cacheInfo = re.ReplaceAllString(cacheInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			cacheInfoArray := re.FindAllString(cacheInfo, -1)

			var subCache *CacheInfo = new(CacheInfo)

			for _, subCacheInfo := range cacheInfoArray {
				subCacheInfoArray := strings.Split(subCacheInfo, ":")
				if len(subCacheInfoArray) == 2 {
					key := strings.TrimSpace(subCacheInfoArray[0])
					value := strings.TrimSpace(subCacheInfoArray[1])
					switch key {
					case "Socket Designation":
						subCache.SocketDesignation = value
					case "Configuration":
						subCache.Configuration = value
					case "Operational Mode":
						subCache.OperationalMode = value
					case "Location":
						subCache.Location = value
					case "Installed Size":
						subCache.InstalledSize = value
					case "Maximum Size":
						subCache.MaximumSize = value
					case "Supported SRAM Types":
						flagsArray := strings.Split(value, "|")
						for index, subValue := range flagsArray {
							flagsArray[index] = strings.TrimSpace(subValue)
						}
						subCache.SupportedSRAMTypes = flagsArray
					case "Installed SRAM Type":
						subCache.InstalledSRAMType = value
					case "Speed":
						subCache.Speed = value
					case "Error Correction Type":
						subCache.ErrorCorrectionType = value
					case "System Type":
						subCache.SystemType = value
					case "Associativity":
						subCache.Associativity = value
					}
				}
			}
			result = append(result, subCache)
		}
	}
	return result, nil
}

// dmidecode -t connector
type PortConnectorInfo struct {
	//Internal Reference Designator: Not Available
	InternalReferenceDesignator string
	//Internal Connector Type: None
	InternalConnectorType string
	//External Reference Designator: External Monitor
	ExternalReferenceDesignator string
	//External Connector Type: DB-15 female
	ExternalConnectorType string
	//Port Type: Video Port
	PortType string
}

func (d *DmiDecode) QueryConnector() ([]*PortConnectorInfo, error) {
	cmd := fmt.Sprintf("%s -t connector", d.Path)
	if DEBUG {
		log.Println("now query connector info: " + cmd)
	}
	connector, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result = make([]*PortConnectorInfo, 0)
	connectorArray := strings.Split(connector, "\n\n")
	for _, connectorInfo := range connectorArray {
		if strings.Contains(connectorInfo, "\nPort Connector Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			connectorInfo = re.ReplaceAllString(connectorInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			connectorInfoArray := re.FindAllString(connectorInfo, -1)

			var subConnector *PortConnectorInfo = new(PortConnectorInfo)
			for _, subConnectorInfo := range connectorInfoArray {
				subConnectorInfoArray := strings.Split(subConnectorInfo, ":")
				if len(subConnectorInfoArray) == 2 {
					key := strings.TrimSpace(subConnectorInfoArray[0])
					value := strings.TrimSpace(subConnectorInfoArray[1])
					switch key {
					case "Internal Reference Designator":
						subConnector.InternalReferenceDesignator = value
					case "Internal Connector Type":
						subConnector.InternalConnectorType = value
					case "External Reference Designator":
						subConnector.ExternalReferenceDesignator = value
					case "External Connector Type":
						subConnector.ExternalConnectorType = value
					case "Port Type":
						subConnector.PortType = value
					}
				}
			}
			result = append(result, subConnector)
		}
	}
	return result, nil
}

// dmidecode -t slot
type SystemSlotInfo struct {
	//Designation: ExpressCard Slot
	Designation string
	//Type: x1 PCI Express
	Type string
	//Current Usage: Available
	CurrentUsage string
	//Length: Other
	Length string
	//ID: 1
	ID string
	//Characteristics:Hot-plug devices are supported
	Characteristics []string
	//Bus Address: 0000:00:00.0
	BusAddress string
}

func (d *DmiDecode) QuerySlot() ([]*SystemSlotInfo, error) {
	cmd := fmt.Sprintf("%s -t slot", d.Path)
	if DEBUG {
		log.Println("now query slot info: " + cmd)
	}
	slot, err := ExecuteCommand(cmd)
	if err != nil {
		if DEBUG {
			log.Println(err)
		}
		return nil, err
	}
	var result = make([]*SystemSlotInfo, 0)
	slotArray := strings.Split(slot, "\n\n")
	for _, slotInfo := range slotArray {
		if strings.Contains(slotInfo, "\nSystem Slot Information\n") {
			re, _ := regexp.Compile("\n\t\t")
			slotInfo = re.ReplaceAllString(slotInfo, "|")
			re, _ = regexp.Compile("\n\t([^\n\t].*)")
			slotInfoArray := re.FindAllString(slotInfo, -1)

			var subSlot *SystemSlotInfo = new(SystemSlotInfo)
			for _, subSlotInfo := range slotInfoArray {
				subSlotInfoArray := strings.Split(subSlotInfo, ":")
				if len(subSlotInfoArray) == 2 {
					key := strings.TrimSpace(subSlotInfoArray[0])
					value := strings.TrimSpace(subSlotInfoArray[1])
					switch key {
					case "Designation":
						subSlot.Designation = value
					case "Type":
						subSlot.Type = value
					case "Current Usage":
						subSlot.CurrentUsage = value
					case "Length":
						subSlot.Length = value
					case "Characteristics":
						characterArray := strings.Split(value, "|")
						for index, subValue := range characterArray {
							characterArray[index] = strings.TrimSpace(subValue)
						}
						subSlot.Characteristics = characterArray
					case "Bus Address":
						subSlot.BusAddress = value
					}
				}
			}
			result = append(result, subSlot)
		}
	}
	return result, nil
}
