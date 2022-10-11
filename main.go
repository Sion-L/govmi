package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vapi/rest"
	"github.com/vmware/govmomi/vapi/vcenter"
	"govmi/utils"
	"os"
)

var (
	template  string
	hostIP    string
	hostname  string
	GuestHost string
	GuestName string
)

func Master(ctx context.Context, name, hostIP, hostname, networkKey string, stores []*object.Datastore, networks []object.NetworkReference, resourcePools []*object.ResourcePool, folders []*object.Folder, item *library.Item, client *govmomi.Client, rc *rest.Client) error {
	deploy := vcenter.Deploy{
		DeploymentSpec: vcenter.DeploymentSpec{
			Name:               name,
			DefaultDatastoreID: stores[7].Reference().Value,
			AcceptAllEULA:      true,
			NetworkMappings: []vcenter.NetworkMapping{{
				Key:   networkKey,
				Value: networks[11].Reference().Value,
			}},
			StorageMappings: []vcenter.StorageMapping{{
				Key: "",
				Value: vcenter.StorageGroupMapping{
					Type:         "DATASTORE",
					DatastoreID:  stores[7].Reference().Value,
					Provisioning: "thin",
				},
			}},
			StorageProvisioning: "thin",
		},
		Target: vcenter.Target{
			ResourcePoolID: resourcePools[13].Reference().Value,
			FolderID:       folders[1].Reference().Value,
		},
	}

	ref, err := vcenter.NewManager(rc).DeployLibraryItem(ctx, item.ID, deploy)
	if err != nil {
		return fmt.Errorf("deploy vm from library failed, %s", err.Error())
	}

	f := find.NewFinder(client.Client)
	obj, err := f.ObjectReference(ctx, *ref)
	if err != nil {
		return fmt.Errorf("Find vm failed, %v\n", err)
	}
	vm := obj.(*object.VirtualMachine)

	// 设置ip
	ip := &utils.IpAddr{
		IP:       hostIP,
		NetMask:  "255.255.254.0",
		Gateway:  "192.168.110.1",
		HostName: hostname,
		DNS:      "192.168.110.244",
	}
	err = ip.SetIP(ctx, vm)
	if err != nil {
		return err
	}

	//开机
	err = utils.PowerOn(ctx, vm)
	return err
}

func Slave(ctx context.Context, name, hostIP, hostname, networkKey string, stores []*object.Datastore, networks []object.NetworkReference, resourcePools []*object.ResourcePool, folders []*object.Folder, item *library.Item, client *govmomi.Client, rc *rest.Client) error {
	deploy := vcenter.Deploy{
		DeploymentSpec: vcenter.DeploymentSpec{
			Name:               name,
			DefaultDatastoreID: stores[3].Reference().Value,
			AcceptAllEULA:      true,
			NetworkMappings: []vcenter.NetworkMapping{{
				Key:   networkKey,
				Value: networks[11].Reference().Value,
			}},
			StorageMappings: []vcenter.StorageMapping{{
				Key: "",
				Value: vcenter.StorageGroupMapping{
					Type:         "DATASTORE",
					DatastoreID:  stores[3].Reference().Value,
					Provisioning: "thin",
				},
			}},
			StorageProvisioning: "thin",
		},
		Target: vcenter.Target{
			ResourcePoolID: resourcePools[8].Reference().Value,
			FolderID:       folders[1].Reference().Value,
		},
	}

	ref, err := vcenter.NewManager(rc).DeployLibraryItem(ctx, item.ID, deploy)
	if err != nil {
		return fmt.Errorf("deploy vm from library failed, %s", err.Error())
	}

	f := find.NewFinder(client.Client)
	obj, err := f.ObjectReference(ctx, *ref)
	if err != nil {
		return fmt.Errorf("Find vm failed, %v\n", err)
	}
	vm := obj.(*object.VirtualMachine)

	// 设置ip
	ip := &utils.IpAddr{
		IP:       hostIP,
		NetMask:  "255.255.254.0",
		Gateway:  "192.168.110.1",
		HostName: hostname,
		DNS:      "192.168.110.244",
	}
	err = ip.SetIP(ctx, vm)
	if err != nil {
		return err
	}

	//开机
	err = utils.PowerOn(ctx, vm)
	return err
}

func main() {
	flag.StringVar(&template, "template", "agentBug", "./main.exe -template agentBug -host 191 -name xxx -ip 192.168.110.234 -hostname xxx")
	flag.StringVar(&GuestHost, "host", "", "")
	flag.StringVar(&hostIP, "ip", "", "")
	flag.StringVar(&hostname, "hostname", "", "")
	flag.StringVar(&GuestName, "name", "", "")
	flag.Parse()
	if len(os.Args) == 1 {
		os.Exit(0)
	}
	stdout := fmt.Sprintf("正在创建虚拟机...IP: %s\n掩码: %s\n网关: %s\nDNS: %s\n主机名: %s\n", hostIP, "255.255.254.0", "192.168.110.1", "192.168.110.244", hostname)
	fmt.Println(stdout)

	ctx, client, rc := utils.ConnVm()
	resourcePools := utils.GetResourcePools(ctx, client)  // 资源池 : 13 -> /firecloud/host/192.168.110.191/Resources || 8 -> /firecloud/host/192.168.111.3/Resources
	dataStores := utils.GetDataStores(ctx, client)        // 存储 : 7 -> /firecloud/datastore/191  || 3 -> /firecloud/datastore/datastore1
	networks := utils.GetNetwork(ctx, client)             // 网络 : 11 -> /firecloud/network/VM Network
	folders := utils.GetFolders(ctx, client)              // 文件夹 : 1 -> /firecloud/vm/Discovered virtual machine
	items, _ := utils.GetLibraryItem(ctx, rc, "template") // 获取对应资源 : item.Name -> agentBug

	m := vcenter.NewManager(rc)
	fr := vcenter.FilterRequest{Target: vcenter.Target{
		ResourcePoolID: resourcePools[13].Reference().Value,
		FolderID:       folders[1].Reference().Value,
	}}
	r, _ := m.FilterLibraryItem(ctx, items.ID, fr)
	networkKey := r.Networks[0]

	if GuestHost == "191" {
		err := Master(ctx, GuestName, hostIP, hostname, networkKey, dataStores, networks, resourcePools, folders, items, client, rc)
		if err != nil {
			fmt.Println(err)
		}
	}

	if GuestHost == "3" {
		err := Slave(ctx, GuestName, hostIP, hostname, networkKey, dataStores, networks, resourcePools, folders, items, client, rc)
		if err != nil {
			fmt.Println(err)
		}
	}
}
