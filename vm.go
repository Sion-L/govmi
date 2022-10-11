//package main
//
//import (
//	"fmt"
//	"govmi/utils"
//)
//
//func main() {
//	ctx, client, _ := utils.ConnVm()
//	//resource := utils.GetResourcePools(ctx, client)
//	//for k, v := range resource {
//	//	fmt.Println(k, ":", v.InventoryPath)
//	//}
//
//	//dataStore := utils.GetDataStores(ctx, client)
//	//for k, v := range dataStore {
//	//	fmt.Println(k, ":", v.Name())
//	//}
//	//
//	//folder := utils.GetFolders(ctx, client)
//	//for k, v := range folder {
//	//	fmt.Println(k, ":", v.InventoryPath)
//	//}
//	//
//	//networks := utils.GetNetwork(ctx, client)
//	//for k, v := range networks {
//	//	fmt.Println(k, ":", v.GetInventoryPath())
//	//}
//
//	vms := utils.GetVms(ctx, client.Client)
//	var IPlist []string
//	for _, v := range vms {
//		if v.Summary.Guest.IpAddress != "" {
//			fmt.Println(v.Summary.Config.Name, ":", v.Summary.Guest.IpAddress)
//			IPlist = append(IPlist, v.Summary.Config.Name)
//		}
//	}
//	fmt.Println(len(IPlist))
//
//}

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"net/url"
	"os"
)

var client *vim25.Client
var ctx = context.Background()

const (
	VSPHERE_IP       = "vc.firecloud.wan"
	VSPHERE_USERNAME = "administrator@vsphere.local"
	VSPHERE_PASSWORD = "Firecloud123!@#"
	Insecure         = true
)

// NewClient 链接vmware
func NewClient() *vim25.Client {

	u := &url.URL{
		Scheme: "https",
		Host:   VSPHERE_IP,
		Path:   "/sdk",
	}

	u.User = url.UserPassword(VSPHERE_USERNAME, VSPHERE_PASSWORD)
	client, err := govmomi.NewClient(ctx, u, Insecure)
	if err != nil {
		panic(err)
	}
	return client.Client
}

// VmsHost 主机结构体
type VmsHost struct {
	Name string
	Ip   string
}

// VmsHosts 主机列表结构体
type VmsHosts struct {
	VmsHosts []VmsHost
}

// NewVmsHosts 初始化结构体
func NewVmsHosts() *VmsHosts {
	return &VmsHosts{
		VmsHosts: make([]VmsHost, 200),
	}
}

// 虚拟机表
type Vm struct {
	Uuid       string
	Vc         string
	Esxi       string
	Name       string
	Ip         string
	PowerState string
}

// AddHost 新增主机
func (vmshosts *VmsHosts) AddHost(name string, ip string) {
	host := &VmsHost{name, ip}
	vmshosts.VmsHosts = append(vmshosts.VmsHosts, *host)
}

// SelectHost 查询主机ip
func (vmshosts *VmsHosts) SelectHost(name string) string {
	ip := "None"
	for _, hosts := range vmshosts.VmsHosts {
		if hosts.Name == name {
			ip = hosts.Ip
		}
	}
	return ip
}

// GetHosts 读取主机信息
func GetHosts(client *vim25.Client, vmshosts *VmsHosts) {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"HostSystem"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var hss []mo.HostSystem
	err = v.Retrieve(ctx, []string{"HostSystem"}, []string{"summary"}, &hss)
	if err != nil {
		panic(err)
	}
	// fmt.Printf("主机名:\t%s\n", hss[0].Summary.Host.Value)
	// fmt.Printf("IP:\t%s\n", hss[0].Summary.Config.Name)
	for _, hs := range hss {
		vmshosts.AddHost(hs.Summary.Host.Value, hs.Summary.Config.Name)
	}
}

// GetVms获取所有vm信息
func GetVms(client *vim25.Client, vmshosts *VmsHosts) []*Vm {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		panic(err)
	}
	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "runtime", "datastore"}, &vms)
	if err != nil {
		panic(err)
	}
	// 输出虚拟机信息到csv
	file, _ := os.OpenFile("./vms.csv", os.O_WRONLY|os.O_CREATE, os.ModePerm)
	//防止中文乱码
	file.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(file)
	w.Write([]string{"宿主机", "虚拟机", "系统", "IP地址"})
	w.Flush()
	for _, vm := range vms {
		//虚拟机资源信息
		//res := strconv.Itoa(int(vm.Summary.Config.MemorySizeMB)) + " MB " + strconv.Itoa(int(vm.Summary.Config.NumCpu)) + " vCPU(s) " + units.ByteSize(vm.Summary.Storage.Committed+vm.Summary.Storage.Uncommitted).String()
		//w.Write([]string{vmshosts.SelectHost(vm.Summary.Runtime.Host.Value), vm.Summary.Config.Name, vm.Summary.Config.GuestFullName, string(vm.Summary.Runtime.PowerState), vm.Summary.Guest.IpAddress, res})
		w.Write([]string{vm.Summary.Config.Name, vm.Summary.Config.GuestFullName, vm.Summary.Guest.IpAddress, vm.Summary.Guest.IpAddress})
		w.Flush()
	}
	file.Close()

	// 批量插入到数据库
	var modelVms []*Vm
	for _, vm := range vms {
		modelVms = append(modelVms, &Vm{
			Esxi: vm.Summary.Runtime.Host.Value,
			Name: vm.Summary.Config.Name,
			Ip:   vm.Summary.Guest.IpAddress,
		})
	}
	return modelVms
}

func main() {
	vms := GetVms(NewClient(), NewVmsHosts())
	//var hostlist []string
	for _, v := range vms {
		if v.Ip == "" {
			fmt.Println(v.Name)
		}
	}
}
