/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"fmt"
	"path/filepath"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/apis/experimental"
	"k8s.io/kubernetes/pkg/util"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("ConfigData", func() {
	f := NewFramework("configData")

	It("should be consumable as a mounted volume", func() {
		name := "config-data-test-" + string(util.NewUUID())
		volumeName := "config-data-volume"
		volumeMountPath := filepath.Join("/etc", volumeName)

		cfg := &experimental.ConfigData{
			ObjectMeta: api.ObjectMeta{
				Namespace: f.NameSpace.Name,
				Name:      name,
			},
			Data: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			},
		}

		By(fmt.Sprintf("Creating ConfigData with name %s", cfg.Name))
		defer func() {
			By("Cleaning up the ConfigData")
			if err := f.Client.ConfigData(f.Namespace.Name).Delete(cfg.Name); err != nil {
				Failf("unable to delete ConfigData %v: %v", cfg.Name, err)
			}
		}()

		cfg, err := f.Client.ConfigData(f.Namespace.Name).Create(cfg)
		if err != nil {
			Failf("unable to create test ConfigData %v: %v", cfg.Name, err)
		}

		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name: "pod-config-data-" + string(util.NewUUID()),
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Name:    "client-container",
						Image:   "gcr.io/google_containers/busybox",
						Command: []string{"sh", "-c", "cat /etc/foo1.bar /etc/foo2.bar"},
						VolumeMounts: []api.VolumeMount{
							{
								Name:      volumeName,
								MountPath: volumeMountPath,
								ReadOnly:  true,
							},
						},
					},
				},
				Volumes: []api.Volume{
					{
						Name: volumeName,
						VolumeSource: api.VolumeSource{
							DownwardAPI: &api.DownwardAPIVolumeSource{
								Items: []DownwardAPIVolumeFile{
									{
										Path: "foo1.bar",
										ConfigDataRef: api.ConfigDataSelector{
											APIVersion:     "v1",
											ConfigDataName: name,
											ConfigDataKey:  "key-1",
										},
									},
									{
										Path: "foo2.bar",
										ConfigDataRef: api.ConfigDataSelector{
											APIVersion:     "v1",
											ConfigDataName: name,
											ConfigDataKey:  "key-2",
										},
									},
								},
							},
						},
					},
				},
				RestartPolicy: api.RestartPolicyNever,
			},
		}

		testContainerOutputInNamespace("consume ConfigData as a mounted volume", f.Client, pod, 0, []string{
			fmt.Sprintf("key-1=\"value-1\"\n"),
			fmt.Sprintf("key-2=\"value-2\"\n"),
		}, f.Namespace.Name)
	})

	It("should be consumable as environment variables", func() {
		name := "config-data-test-" + string(util.NewUUID())
		volumeName := "config-data-volume"
		volumeMountPath := filepath.Join("/etc", volumeName)

		cfg := &experimental.ConfigData{
			ObjectMeta: api.ObjectMeta{
				Namespace: f.NameSpace.Name,
				Name:      name,
			},
			Data: map[string]string{
				"key-1": "value-1",
				"key-2": "value-2",
			},
		}

		By(fmt.Sprintf("Creating ConfigData with name %s", cfg.Name))
		defer func() {
			By("Cleaning up the ConfigData")
			if err := f.Client.ConfigData(f.Namespace.Name).Delete(cfg.Name); err != nil {
				Failf("unable to delete ConfigData %v: %v", cfg.Name, err)
			}
		}()

		cfg, err := f.Client.ConfigData(f.Namespace.Name).Create(cfg)
		if err != nil {
			Failf("unable to create test ConfigData %v: %v", cfg.Name, err)
		}

		pod := &api.Pod{
			ObjectMeta: api.ObjectMeta{
				Name: "pod-config-data-" + string(util.NewUUID()),
			},
			Spec: api.PodSpec{
				Containers: []api.Container{
					{
						Name:    "client-container",
						Image:   "gcr.io/google_containers/busybox",
						Command: []string{"sh", "-c", "env"},
						Env: []api.EnvVar{
							{
								Name: "FOO_BAR_1",
								ValueFrom: &api.EnvVarSource{
									ConfigDataRef: api.ConfigDataSelector{
										APIVersion:     "v1",
										ConfigDataName: name,
										ConfigDataKey:  "key-1",
									},
								},
							},
							{
								Name: "FOO_BAR_2",
								ValueFrom: &api.EnvVarSource{
									ConfigDataRef: api.ConfigDataSelector{
										APIVersion:     "v1",
										ConfigDataName: name,
										ConfigDataKey:  "key-2",
									},
								},
							},
						},
					},
				},
				RestartPolicy: api.RestartPolicyNever,
			},
		}

		testContainerOutputInNamespace("consume ConfigData as env vars", f.Client, pod, 0, []string{
			fmt.Sprintf("FOO_BAR_1=\"value-1\"\n"),
			fmt.Sprintf("FOO_BAR_2=\"value-2\"\n"),
		}, f.Namespace.Name)
	})
})
