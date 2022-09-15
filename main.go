package main

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"
	"os"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	kv1 "kubevirt.io/api/core/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func createConfigMap(ctx context.Context, client ctrlclient.Client) (*corev1.ConfigMap, error) {
	cm := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sysprep-config-vm",
			Namespace: "default",
		},
		Data: map[string]string{
			"foo": "bar",
		},
	}
	return &cm, client.Create(ctx, &cm)
}

func createVM(ctx context.Context, client ctrlclient.Client, cmName string) error {
	vm := kv1.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-vm",
			Namespace: "default",
			Labels: map[string]string{
				"kubevirt.io/vm": "test-vm",
			},
		},
		Spec: kv1.VirtualMachineSpec{
			Running: pointer.Bool(true),
			Template: &kv1.VirtualMachineInstanceTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"kubevirt.io/vm": "test-vm",
					},
				},
				Spec: kv1.VirtualMachineInstanceSpec{
					Domain: kv1.DomainSpec{
						Resources: kv1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("1"),
								corev1.ResourceMemory: resource.MustParse("2Gi"),
							},
						},
						Devices: kv1.Devices{
							Disks: []kv1.Disk{
								{
									Name: "primary",
									DiskDevice: kv1.DiskDevice{
										Disk: &kv1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
								{
									Name: "config-map",
									DiskDevice: kv1.DiskDevice{
										Disk: &kv1.DiskTarget{
											Bus: "virtio",
										},
									},
								},
							},
						},
					},
					Volumes: []kv1.Volume{
						{
							Name: "primary",
							VolumeSource: kv1.VolumeSource{
								ContainerDisk: &kv1.ContainerDiskSource{
									Image: "docker.io/kubevirt/cirros-container-disk-demo:devel",
								},
							},
						},
						{
							Name: "config-map",
							VolumeSource: kv1.VolumeSource{
								ConfigMap: &kv1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cmName,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return client.Create(ctx, &vm)
}

func createCMFromYaml(ctx context.Context, client ctrlclient.Client, yamlPath string) error {
	cm := &corev1.ConfigMap{}

	f, err := os.Open(yamlPath)
	if err != nil {
		return err
	}

	if err = yaml.NewYAMLToJSONDecoder(f).Decode(cm); err != nil {
		return err
	}

	if err = client.Create(ctx, cm); err != nil {
		return err
	}

	return nil
}

func createVMFromYaml(ctx context.Context, client ctrlclient.Client, yamlPath string) error {
	vm := &kv1.VirtualMachine{}

	f, err := os.Open(yamlPath)
	if err != nil {
		return err
	}

	if err = yaml.NewYAMLToJSONDecoder(f).Decode(vm); err != nil {
		return err
	}

	if err = client.Create(ctx, vm); err != nil {
		return err
	}

	return nil
}

func createSvcFromYaml(ctx context.Context, client ctrlclient.Client, yamlPath string) error {
	svc := &corev1.Service{}

	f, err := os.Open(yamlPath)
	if err != nil {
		return err
	}

	if err = yaml.NewYAMLToJSONDecoder(f).Decode(svc); err != nil {
		return err
	}

	if err = client.Create(ctx, svc); err != nil {
		return err
	}

	return nil
}

func main() {
	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	kv1.AddToScheme(scheme)

	ctx := context.TODO()

	// controller-runtime already registers a kubeconfig flag in the init function
	restConfig := ctrl.GetConfigOrDie()
	client, err := ctrlclient.New(restConfig, ctrlclient.Options{
		Scheme: scheme,
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err = createCMFromYaml(ctx, client, "cm.yaml"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Config map created")

	if err = createVMFromYaml(ctx, client, "vm.yaml"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Virtual Machine created")

	if err = createSvcFromYaml(ctx, client, "svc.yaml"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Service created")

	// The below code creates k8s/kubevirt resources based on the created resource structure in code

	//_, err = createConfigMap(ctx, client)
	//if err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}

	//if err = createVM(ctx, client, cm.Name); err != nil {
	//	fmt.Println(err)
	//	os.Exit(1)
	//}
}
