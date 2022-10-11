package utils

import (
	"context"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
	"log"
	"net/url"
	"os"
)

const (
	ip              = "vc.firecloud.wan"
	libraryName     = "template"
	libraryItemType = "ovf"
	user            = "administrator@vsphere.local"
	password        = "Firecloud123!@#"
)

var itemList *library.Item

type IpAddr struct {
	IP       string
	NetMask  string
	Gateway  string
	HostName string
	DNS      string
}

// 连接
func ConnVm() (context.Context, *govmomi.Client, *rest.Client) {
	u := &url.URL{
		Scheme: "https",
		Host:   ip,
		Path:   "/sdk",
	}
	ctx := context.Background()
	u.User = url.UserPassword(user, password)
	client, err := govmomi.NewClient(ctx, u, true)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Login to vsphere failed, %v", err)
		os.Exit(1)
	}

	rc := rest.NewClient(client.Client)
	if err := rc.Login(ctx, url.UserPassword(user, password)); err != nil {
		fmt.Fprintf(os.Stderr, "rc.Login failed, %v", err)
		os.Exit(1)
	}
	return ctx, client, rc
}

func GetResourcePools(ctx context.Context, client *govmomi.Client) []*object.ResourcePool {
	finder := find.NewFinder(client.Client)
	resourcePools, err := finder.ResourcePoolList(ctx, "*") // resourcePools[0]是191
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list resource pool at vc %s, %v", ip, err)
		os.Exit(1)
	}
	return resourcePools
}

func GetDataStores(ctx context.Context, client *govmomi.Client) []*object.Datastore {
	finder := find.NewFinder(client.Client)
	datastores, err := finder.DatastoreList(ctx, "*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list datastore at vc %s, %v", ip, err)
		os.Exit(1)
	}
	return datastores
}

func GetNetwork(ctx context.Context, client *govmomi.Client) []object.NetworkReference {
	finder := find.NewFinder(client.Client)
	networks, err := finder.NetworkList(ctx, "*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list network at vc %s, %v", ip, err)
		os.Exit(1)
	}
	return networks
}

func GetFolders(ctx context.Context, client *govmomi.Client) []*object.Folder {
	finder := find.NewFinder(client.Client)
	folders, err := finder.FolderList(ctx, "*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list folder at vc %s, %v", ip, err)
		os.Exit(1)
	}
	return folders
}

func GetLibraryItem(ctx context.Context, rc *rest.Client, template string) (*library.Item, error) {
	if template == "template" {
		libraryItemName := "agentBug"
		m := library.NewManager(rc)
		libraries, _ := m.FindLibrary(ctx, library.Find{Name: libraryName})
		items, _ := m.FindLibraryItems(ctx, library.FindItem{Name: libraryItemName,
			Type: libraryItemType, LibraryID: libraries[0]})
		item, err := m.GetLibraryItem(ctx, items[0])
		if err != nil {
			fmt.Printf("Get library item by %s failed, %v", items[0], err)
			return nil, err
		}
		itemList = item
	}
	return itemList, nil
}

// 修改cpu
func SetCPUAndMem(ctx context.Context, vm *object.VirtualMachine, cpuNum int32, mem int64) error {
	spec := types.VirtualMachineConfigSpec{
		NumCPUs:             cpuNum,
		NumCoresPerSocket:   cpuNum / 2,
		MemoryMB:            1024 * mem,
		CpuHotAddEnabled:    types.NewBool(true),
		MemoryHotAddEnabled: types.NewBool(true),
	}
	task, err := vm.Reconfigure(ctx, spec)
	if err != nil {
		return err
	}
	return task.Wait(ctx)
}

// 修改ip
func (p *IpAddr) SetIP(ctx context.Context, vm *object.VirtualMachine) error {
	customSpec := types.CustomizationSpec{
		NicSettingMap: []types.CustomizationAdapterMapping{
			types.CustomizationAdapterMapping{
				Adapter: types.CustomizationIPSettings{
					Ip: &types.CustomizationFixedIp{
						IpAddress: p.IP,
					},
					SubnetMask:    p.NetMask,
					Gateway:       []string{p.Gateway},
					DnsServerList: []string{p.DNS},
				},
			},
		},
		Identity: &types.CustomizationLinuxPrep{
			HostName: &types.CustomizationFixedName{
				Name: p.HostName,
			},
			HwClockUTC: types.NewBool(true),
		},
		GlobalIPSettings: types.CustomizationGlobalIPSettings{
			DnsServerList: []string{p.DNS},
		},
	}
	task, err := vm.Customize(ctx, customSpec)
	if err != nil {
		return err
	}
	return task.Wait(ctx)
}

// 开机
func PowerOn(ctx context.Context, vm *object.VirtualMachine) error {
	task, err := vm.PowerOn(ctx)
	if err != nil {
		log.Fatalf("Failed to power on %s", vm.Name())
		return err
	}
	return task.Wait(ctx)
}

func GetVms(ctx context.Context, client *vim25.Client) []mo.VirtualMachine {
	m := view.NewManager(client)
	v, err := m.CreateContainerView(ctx, client.ServiceContent.RootFolder, []string{"VirtualMachine"}, true)
	if err != nil {
		fmt.Println(err)
	}

	defer v.Destroy(ctx)
	var vms []mo.VirtualMachine
	err = v.Retrieve(ctx, []string{"VirtualMachine"}, []string{"summary", "runtime", "datastore"}, &vms)
	if err != nil {
		fmt.Println(err)
	}

	return vms
}
