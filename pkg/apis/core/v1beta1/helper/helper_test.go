// Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper_test

import (
	"time"

	"github.com/Masterminds/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1beta1constants "github.com/gardener/gardener/pkg/apis/core/v1beta1/constants"
	. "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
)

var _ = Describe("helper", func() {
	var (
		trueVar                 = true
		falseVar                = false
		expirationDateInThePast = metav1.Time{Time: time.Now().AddDate(0, 0, -1)}
	)

	Describe("errors", func() {
		var zeroTime metav1.Time

		DescribeTable("#UpdatedCondition",
			func(condition gardencorev1beta1.Condition, status gardencorev1beta1.ConditionStatus, reason, message string, codes []gardencorev1beta1.ErrorCode, matcher types.GomegaMatcher) {
				updated := UpdatedCondition(condition, status, reason, message, codes...)

				Expect(updated).To(matcher)
			},
			Entry("no update",
				gardencorev1beta1.Condition{
					Status:  gardencorev1beta1.ConditionTrue,
					Reason:  "reason",
					Message: "message",
				},
				gardencorev1beta1.ConditionTrue,
				"reason",
				"message",
				nil,
				MatchFields(IgnoreExtras, Fields{
					"Status":             Equal(gardencorev1beta1.ConditionTrue),
					"Reason":             Equal("reason"),
					"Message":            Equal("message"),
					"LastTransitionTime": Equal(zeroTime),
					"LastUpdateTime":     Not(Equal(zeroTime)),
				}),
			),
			Entry("update reason",
				gardencorev1beta1.Condition{
					Status:  gardencorev1beta1.ConditionTrue,
					Reason:  "reason",
					Message: "message",
				},
				gardencorev1beta1.ConditionTrue,
				"OtherReason",
				"message",
				nil,
				MatchFields(IgnoreExtras, Fields{
					"Status":             Equal(gardencorev1beta1.ConditionTrue),
					"Reason":             Equal("OtherReason"),
					"Message":            Equal("message"),
					"LastTransitionTime": Equal(zeroTime),
					"LastUpdateTime":     Not(Equal(zeroTime)),
				}),
			),
			Entry("update codes",
				gardencorev1beta1.Condition{
					Status:  gardencorev1beta1.ConditionTrue,
					Reason:  "reason",
					Message: "message",
				},
				gardencorev1beta1.ConditionTrue,
				"OtherReason",
				"message",
				[]gardencorev1beta1.ErrorCode{gardencorev1beta1.ErrorCleanupClusterResources},
				MatchFields(IgnoreExtras, Fields{
					"Status":             Equal(gardencorev1beta1.ConditionTrue),
					"Reason":             Equal("OtherReason"),
					"Message":            Equal("message"),
					"Codes":              Equal([]gardencorev1beta1.ErrorCode{gardencorev1beta1.ErrorCleanupClusterResources}),
					"LastTransitionTime": Equal(zeroTime),
					"LastUpdateTime":     Not(Equal(zeroTime)),
				}),
			),
			Entry("update status",
				gardencorev1beta1.Condition{
					Status:  gardencorev1beta1.ConditionTrue,
					Reason:  "reason",
					Message: "message",
				},
				gardencorev1beta1.ConditionFalse,
				"OtherReason",
				"message",
				nil,
				MatchFields(IgnoreExtras, Fields{
					"Status":             Equal(gardencorev1beta1.ConditionFalse),
					"Reason":             Equal("OtherReason"),
					"Message":            Equal("message"),
					"LastTransitionTime": Not(Equal(zeroTime)),
					"LastUpdateTime":     Not(Equal(zeroTime)),
				}),
			),
		)

		Describe("#MergeConditions", func() {
			It("should merge the conditions", func() {
				var (
					typeFoo gardencorev1beta1.ConditionType = "foo"
					typeBar gardencorev1beta1.ConditionType = "bar"
				)

				oldConditions := []gardencorev1beta1.Condition{
					{
						Type:   typeFoo,
						Reason: "hugo",
					},
				}

				result := MergeConditions(oldConditions, gardencorev1beta1.Condition{Type: typeFoo}, gardencorev1beta1.Condition{Type: typeBar})

				Expect(result).To(Equal([]gardencorev1beta1.Condition{{Type: typeFoo}, {Type: typeBar}}))
			})
		})

		Describe("#GetCondition", func() {
			It("should return the found condition", func() {
				var (
					conditionType gardencorev1beta1.ConditionType = "test-1"
					condition                                     = gardencorev1beta1.Condition{
						Type: conditionType,
					}
					conditions = []gardencorev1beta1.Condition{condition}
				)

				cond := GetCondition(conditions, conditionType)

				Expect(cond).NotTo(BeNil())
				Expect(*cond).To(Equal(condition))
			})

			It("should return nil because the required condition could not be found", func() {
				var (
					conditionType gardencorev1beta1.ConditionType = "test-1"
					conditions                                    = []gardencorev1beta1.Condition{}
				)

				cond := GetCondition(conditions, conditionType)

				Expect(cond).To(BeNil())
			})
		})

		Describe("#GetOrInitCondition", func() {
			It("should get the existing condition", func() {
				var (
					c          = gardencorev1beta1.Condition{Type: "foo"}
					conditions = []gardencorev1beta1.Condition{c}
				)

				Expect(GetOrInitCondition(conditions, "foo")).To(Equal(c))
			})

			It("should return a new, initialized condition", func() {
				tmp := Now
				Now = func() metav1.Time {
					return metav1.NewTime(time.Unix(0, 0))
				}
				defer func() { Now = tmp }()

				Expect(GetOrInitCondition(nil, "foo")).To(Equal(InitCondition("foo")))
			})
		})

		DescribeTable("#IsResourceSupported",
			func(resources []gardencorev1beta1.ControllerResource, resourceKind, resourceType string, expectation bool) {
				Expect(IsResourceSupported(resources, resourceKind, resourceType)).To(Equal(expectation))
			},
			Entry("expect true",
				[]gardencorev1beta1.ControllerResource{
					{
						Kind: "foo",
						Type: "bar",
					},
				},
				"foo",
				"bar",
				true,
			),
			Entry("expect true",
				[]gardencorev1beta1.ControllerResource{
					{
						Kind: "foo",
						Type: "bar",
					},
				},
				"foo",
				"BAR",
				true,
			),
			Entry("expect false",
				[]gardencorev1beta1.ControllerResource{
					{
						Kind: "foo",
						Type: "bar",
					},
				},
				"foo",
				"baz",
				false,
			),
		)

		DescribeTable("#IsControllerInstallationSuccessful",
			func(conditions []gardencorev1beta1.Condition, expectation bool) {
				controllerInstallation := gardencorev1beta1.ControllerInstallation{
					Status: gardencorev1beta1.ControllerInstallationStatus{
						Conditions: conditions,
					},
				}
				Expect(IsControllerInstallationSuccessful(controllerInstallation)).To(Equal(expectation))
			},
			Entry("expect true",
				[]gardencorev1beta1.Condition{
					{
						Type:   gardencorev1beta1.ControllerInstallationInstalled,
						Status: gardencorev1beta1.ConditionTrue,
					},
					{
						Type:   gardencorev1beta1.ControllerInstallationHealthy,
						Status: gardencorev1beta1.ConditionTrue,
					},
				},
				true,
			),
			Entry("expect false",
				[]gardencorev1beta1.Condition{
					{
						Type:   gardencorev1beta1.ControllerInstallationInstalled,
						Status: gardencorev1beta1.ConditionFalse,
					},
				},
				false,
			),
			Entry("expect false",
				[]gardencorev1beta1.Condition{
					{
						Type:   gardencorev1beta1.ControllerInstallationHealthy,
						Status: gardencorev1beta1.ConditionFalse,
					},
				},
				false,
			),
			Entry("expect false",
				[]gardencorev1beta1.Condition{
					{
						Type:   gardencorev1beta1.ControllerInstallationInstalled,
						Status: gardencorev1beta1.ConditionTrue,
					},
					{
						Type:   gardencorev1beta1.ControllerInstallationHealthy,
						Status: gardencorev1beta1.ConditionFalse,
					},
				},
				false,
			),
			Entry("expect false",
				[]gardencorev1beta1.Condition{
					{
						Type:   gardencorev1beta1.ControllerInstallationInstalled,
						Status: gardencorev1beta1.ConditionFalse,
					},
					{
						Type:   gardencorev1beta1.ControllerInstallationHealthy,
						Status: gardencorev1beta1.ConditionTrue,
					},
				},
				false,
			),
			Entry("expect false",
				[]gardencorev1beta1.Condition{},
				false,
			),
		)

		DescribeTable("#TaintsHave",
			func(taints []gardencorev1beta1.SeedTaint, key string, expectation bool) {
				Expect(TaintsHave(taints, key)).To(Equal(expectation))
			},
			Entry("taint exists", []gardencorev1beta1.SeedTaint{{Key: "foo"}}, "foo", true),
			Entry("taint does not exist", []gardencorev1beta1.SeedTaint{{Key: "foo"}}, "bar", false),
		)

		Describe("#ReadShootedSeed", func() {
			var (
				shoot                    *gardencorev1beta1.Shoot
				defaultReplicas          int32 = 3
				defaultMinReplicas       int32 = 3
				defaultMaxReplicas       int32 = 3
				defaultMinimumVolumeSize       = "20Gi"

				defaultAPIServerAutoscaler = ShootedSeedAPIServerAutoscaler{
					MinReplicas: &defaultMinReplicas,
					MaxReplicas: defaultMaxReplicas,
				}

				defaultAPIServer = ShootedSeedAPIServer{
					Replicas:   &defaultReplicas,
					Autoscaler: &defaultAPIServerAutoscaler,
				}

				defaultShootedSeed = ShootedSeed{
					APIServer: &defaultAPIServer,
					Backup:    &gardencorev1beta1.SeedBackup{},
				}
			)

			BeforeEach(func() {
				shoot = &gardencorev1beta1.Shoot{
					ObjectMeta: metav1.ObjectMeta{
						Namespace:   v1beta1constants.GardenNamespace,
						Annotations: nil,
					},
				}
			})

			It("should return false,nil,nil because shoot is not in the garden namespace", func() {
				shoot.Namespace = "default"

				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(BeNil())
			})

			It("should return false,nil,nil because annotation is not set", func() {
				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(BeNil())
			})

			It("should return false,nil,nil because annotation is set with no usages", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "",
				}

				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(BeNil())
			})

			It("should return true,nil,nil because annotation is set with normal usage", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true",
				}

				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(Equal(&defaultShootedSeed))
			})

			It("should return true,true,true because annotation is set with protected and visible usage", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,protected,visible",
				}

				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(Equal(&ShootedSeed{
					Protected: &trueVar,
					Visible:   &trueVar,
					APIServer: &defaultAPIServer,
					Backup:    &gardencorev1beta1.SeedBackup{},
				}))
			})

			It("should return true,true,true because annotation is set with unprotected and invisible usage", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,unprotected,invisible",
				}

				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(Equal(&ShootedSeed{
					Protected:         &falseVar,
					Visible:           &falseVar,
					APIServer:         &defaultAPIServer,
					Backup:            &gardencorev1beta1.SeedBackup{},
					MinimumVolumeSize: nil,
				}))
			})

			It("should return the min volume size because annotation is set properly", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,unprotected,invisible,minimumVolumeSize=20Gi",
				}

				shootedSeed, err := ReadShootedSeed(shoot)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(Equal(&ShootedSeed{
					Protected:         &falseVar,
					Visible:           &falseVar,
					APIServer:         &defaultAPIServer,
					Backup:            &gardencorev1beta1.SeedBackup{},
					MinimumVolumeSize: &defaultMinimumVolumeSize,
				}))
			})

			It("should return a filled apiserver config", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,apiServer.replicas=1,apiServer.autoscaler.minReplicas=2,apiServer.autoscaler.maxReplicas=3",
				}

				shootedSeed, err := ReadShootedSeed(shoot)

				var (
					one   int32 = 1
					two   int32 = 2
					three int32 = 3
				)

				Expect(err).NotTo(HaveOccurred())
				Expect(shootedSeed).To(Equal(&ShootedSeed{
					APIServer: &ShootedSeedAPIServer{
						Replicas: &one,
						Autoscaler: &ShootedSeedAPIServerAutoscaler{
							MinReplicas: &two,
							MaxReplicas: three,
						},
					},
					Backup: &gardencorev1beta1.SeedBackup{},
				}))
			})

			It("should fail due to maxReplicas not being specified", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,apiServer.autoscaler.minReplicas=2",
				}

				_, err := ReadShootedSeed(shoot)
				Expect(err).To(HaveOccurred())
			})

			It("should fail due to API server replicas being less than one", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,apiServer.replicas=0",
				}

				_, err := ReadShootedSeed(shoot)
				Expect(err).To(HaveOccurred())
			})

			It("should fail due to API server autoscaler minReplicas being less than one", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,apiServer.autoscaler.minReplicas=0,apiServer.autoscaler.maxReplicas=1",
				}

				_, err := ReadShootedSeed(shoot)
				Expect(err).To(HaveOccurred())
			})

			It("should fail due to API server autoscaler maxReplicas being less than one", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,apiServer.autoscaler.maxReplicas=0",
				}

				_, err := ReadShootedSeed(shoot)
				Expect(err).To(HaveOccurred())
			})

			It("should fail due to API server autoscaler minReplicas being greater than maxReplicas", func() {
				shoot.Annotations = map[string]string{
					v1beta1constants.AnnotationShootUseAsSeed: "true,apiServer.autoscaler.maxReplicas=1,apiServer.autoscaler.minReplicas=2",
				}

				_, err := ReadShootedSeed(shoot)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	DescribeTable("#HibernationIsEnabled",
		func(shoot *gardencorev1beta1.Shoot, hibernated bool) {
			Expect(HibernationIsEnabled(shoot)).To(Equal(hibernated))
		},
		Entry("no hibernation section", &gardencorev1beta1.Shoot{}, false),
		Entry("hibernation.enabled = false", &gardencorev1beta1.Shoot{
			Spec: gardencorev1beta1.ShootSpec{
				Hibernation: &gardencorev1beta1.Hibernation{Enabled: &falseVar},
			},
		}, false),
		Entry("hibernation.enabled = true", &gardencorev1beta1.Shoot{
			Spec: gardencorev1beta1.ShootSpec{
				Hibernation: &gardencorev1beta1.Hibernation{Enabled: &trueVar},
			},
		}, true),
	)

	DescribeTable("#ShootWantsClusterAutoscaler",
		func(shoot *gardencorev1beta1.Shoot, wantsAutoscaler bool) {
			actualWantsAutoscaler, err := ShootWantsClusterAutoscaler(shoot)

			Expect(err).NotTo(HaveOccurred())
			Expect(actualWantsAutoscaler).To(Equal(wantsAutoscaler))
		},

		Entry("no workers",
			&gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{},
			},
			false),

		Entry("one worker no difference in auto scaler max and min",
			&gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{
					Provider: gardencorev1beta1.Provider{
						Workers: []gardencorev1beta1.Worker{{Name: "foo"}},
					},
				},
			},
			false),

		Entry("one worker with difference in auto scaler max and min",
			&gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{
					Provider: gardencorev1beta1.Provider{
						Workers: []gardencorev1beta1.Worker{{Name: "foo", Minimum: 1, Maximum: 2}},
					},
				},
			},
			true),
	)

	var (
		unmanagedType = "unmanaged"
		differentType = "foo"
	)

	DescribeTable("#ShootUsesUnmanagedDNS",
		func(dns *gardencorev1beta1.DNS, expectation bool) {
			shoot := &gardencorev1beta1.Shoot{
				Spec: gardencorev1beta1.ShootSpec{
					DNS: dns,
				},
			}
			Expect(ShootUsesUnmanagedDNS(shoot)).To(Equal(expectation))
		},

		Entry("no dns", nil, false),
		Entry("no dns providers", &gardencorev1beta1.DNS{}, false),
		Entry("dns providers but no type", &gardencorev1beta1.DNS{Providers: []gardencorev1beta1.DNSProvider{{}}}, false),
		Entry("dns providers but different type", &gardencorev1beta1.DNS{Providers: []gardencorev1beta1.DNSProvider{{Type: &differentType}}}, false),
		Entry("dns providers and unmanaged type", &gardencorev1beta1.DNS{Providers: []gardencorev1beta1.DNSProvider{{Type: &unmanagedType}}}, true),
	)

	DescribeTable("#IsAPIServerExposureManaged",
		func(obj metav1.Object, expected bool) {
			Expect(IsAPIServerExposureManaged(obj)).To(Equal(expected))
		},
		Entry("object is nil",
			nil,
			false,
		),
		Entry("label is not present",
			&metav1.ObjectMeta{Labels: map[string]string{
				"foo": "bar",
			}},
			false,
		),
		Entry("label's value is not the same",
			&metav1.ObjectMeta{Labels: map[string]string{
				"core.gardener.cloud/apiserver-exposure": "some-dummy-value",
			}},
			false,
		),
		Entry("label's value is gardener-managed",
			&metav1.ObjectMeta{Labels: map[string]string{
				"core.gardener.cloud/apiserver-exposure": "gardener-managed",
			}},
			true,
		),
	)

	DescribeTable("#FindPrimaryDNSProvider",
		func(providers []gardencorev1beta1.DNSProvider, matcher types.GomegaMatcher) {
			Expect(FindPrimaryDNSProvider(providers)).To(matcher)
		},

		Entry("no providers", nil, BeNil()),
		Entry("one non primary provider", []gardencorev1beta1.DNSProvider{
			{Type: pointer.StringPtr("provider")},
		}, Equal(&gardencorev1beta1.DNSProvider{Type: pointer.StringPtr("provider")})),
		Entry("one primary provider", []gardencorev1beta1.DNSProvider{{Type: pointer.StringPtr("provider"),
			Primary: pointer.BoolPtr(true)}}, Equal(&gardencorev1beta1.DNSProvider{Type: pointer.StringPtr("provider"), Primary: pointer.BoolPtr(true)})),
		Entry("multiple w/ one primary provider", []gardencorev1beta1.DNSProvider{
			{
				Type: pointer.StringPtr("provider2"),
			},
			{
				Type:    pointer.StringPtr("provider1"),
				Primary: pointer.BoolPtr(true),
			},
			{
				Type: pointer.StringPtr("provider3"),
			},
		}, Equal(&gardencorev1beta1.DNSProvider{Type: pointer.StringPtr("provider1"), Primary: pointer.BoolPtr(true)})),
		Entry("multiple w/ multiple primary providers", []gardencorev1beta1.DNSProvider{
			{
				Type:    pointer.StringPtr("provider1"),
				Primary: pointer.BoolPtr(true),
			},
			{
				Type:    pointer.StringPtr("provider2"),
				Primary: pointer.BoolPtr(true),
			},
			{
				Type: pointer.StringPtr("provider3"),
			},
		}, Equal(&gardencorev1beta1.DNSProvider{Type: pointer.StringPtr("provider1"), Primary: pointer.BoolPtr(true)})),
	)

	Describe("#ShootMachineImageVersionExists", func() {
		var (
			constraint        gardencorev1beta1.MachineImage
			shootMachineImage gardencorev1beta1.ShootMachineImage
		)

		BeforeEach(func() {
			constraint = gardencorev1beta1.MachineImage{
				Name: "coreos",
				Versions: []gardencorev1beta1.ExpirableVersion{
					{Version: "0.0.2"},
					{Version: "0.0.3"},
				},
			}

			shootMachineImage = gardencorev1beta1.ShootMachineImage{
				Name:    "coreos",
				Version: pointer.StringPtr("0.0.2"),
			}
		})

		It("should determine that the version exists", func() {
			exists, index := ShootMachineImageVersionExists(constraint, shootMachineImage)
			Expect(exists).To(Equal(trueVar))
			Expect(index).To(Equal(0))
		})

		It("should determine that the version does not exist", func() {
			shootMachineImage.Name = "xy"
			exists, _ := ShootMachineImageVersionExists(constraint, shootMachineImage)
			Expect(exists).To(Equal(false))
		})

		It("should determine that the version does not exist", func() {
			shootMachineImage.Version = pointer.StringPtr("0.0.4")
			exists, _ := ShootMachineImageVersionExists(constraint, shootMachineImage)
			Expect(exists).To(Equal(false))
		})
	})

	Describe("Version helper", func() {
		var previewClassification = gardencorev1beta1.ClassificationPreview

		DescribeTable("#GetLatestQualifyingShootMachineImage",
			func(original gardencorev1beta1.MachineImage, expectVersionToBeFound bool, expected *gardencorev1beta1.ShootMachineImage, expectError bool) {
				qualifyingVersionFound, latestVersion, err := GetLatestQualifyingShootMachineImage(original)
				if expectError {
					Expect(err).To(HaveOccurred())
					return
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(qualifyingVersionFound).To(Equal(expectVersionToBeFound))
				Expect(latestVersion).To(Equal(expected))
			},
			Entry("Get latest version",
				gardencorev1beta1.MachineImage{
					Name: "gardenlinux",
					Versions: []gardencorev1beta1.ExpirableVersion{
						{
							Version: "1.17.1",
						},
						{
							Version: "1.15.0",
						},
						{
							Version: "1.14.3",
						},
						{
							Version: "1.13.1",
						},
					},
				},
				true,
				&gardencorev1beta1.ShootMachineImage{
					Name:    "gardenlinux",
					Version: pointer.StringPtr("1.17.1"),
				},
				false,
			),
			Entry("Expect no qualifying version to be found - machine image has only versions in preview and expired versions",
				gardencorev1beta1.MachineImage{
					Name: "gardenlinux",
					Versions: []gardencorev1beta1.ExpirableVersion{
						{
							Version:        "1.17.1",
							Classification: &previewClassification,
						},
						{
							Version:        "1.15.0",
							Classification: &previewClassification,
						},
						{
							Version:        "1.14.3",
							ExpirationDate: &expirationDateInThePast,
						},
						{
							Version:        "1.13.1",
							ExpirationDate: &expirationDateInThePast,
						},
					},
				},
				false,
				nil,
				false,
			),
		)

		DescribeTable("#GetLatestQualifyingVersion",
			func(original []gardencorev1beta1.ExpirableVersion, expectVersionToBeFound bool, expected *gardencorev1beta1.ExpirableVersion, expectError bool) {
				qualifyingVersionFound, latestVersion, err := GetLatestQualifyingVersion(original, nil)
				if expectError {
					Expect(err).To(HaveOccurred())
					return
				}
				Expect(err).ToNot(HaveOccurred())
				Expect(qualifyingVersionFound).To(Equal(expectVersionToBeFound))
				Expect(latestVersion).To(Equal(expected))
			},
			Entry("Get latest non-preview version",
				[]gardencorev1beta1.ExpirableVersion{
					{
						Version:        "1.17.2",
						Classification: &previewClassification,
					},
					{
						Version: "1.17.1",
					},
					{
						Version: "1.15.0",
					},
					{
						Version: "1.14.3",
					},
					{
						Version: "1.13.1",
					},
				},
				true,
				&gardencorev1beta1.ExpirableVersion{
					Version: "1.17.1",
				},
				false,
			),
			Entry("Expect no qualifying version to be found - no latest version could be found",
				[]gardencorev1beta1.ExpirableVersion{},
				false,
				nil,
				false,
			),
			Entry("Expect error, because contains invalid semVer",
				[]gardencorev1beta1.ExpirableVersion{
					{
						Version: "1.213123xx",
					},
				},
				false,
				nil,
				true,
			),
		)

		Describe("#Kubernetes Version Helper", func() {
			DescribeTable("#GetKubernetesVersionForPatchUpdate",
				func(currentVersion string, cloudProfileVersions []gardencorev1beta1.ExpirableVersion, expectedVersion string, qualifyingVersionFound bool) {
					cloudProfile := gardencorev1beta1.CloudProfile{
						Spec: gardencorev1beta1.CloudProfileSpec{
							Kubernetes: gardencorev1beta1.KubernetesSettings{
								Versions: cloudProfileVersions,
							},
						},
					}
					ok, newVersion, err := GetKubernetesVersionForPatchUpdate(&cloudProfile, currentVersion)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(Equal(qualifyingVersionFound))
					Expect(newVersion).To(Equal(expectedVersion))
				},
				Entry("Do not consider preview versions for patch update.",
					"1.12.2",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{
							Version:        "1.12.9",
							Classification: &previewClassification,
						},
						{
							Version:        "1.12.4",
							Classification: &previewClassification,
						},
						// latest qualifying version for updating version 1.12.2
						{Version: "1.12.3"},
						{Version: "1.12.2"},
					},
					"1.12.3",
					true,
				),
				Entry("Do not consider expired versions for patch update.",
					"1.12.2",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{
							Version:        "1.12.9",
							ExpirationDate: &expirationDateInThePast,
						},
						{
							Version:        "1.12.4",
							ExpirationDate: &expirationDateInThePast,
						},
						// latest qualifying version for updating version 1.12.2
						{Version: "1.12.3"},
						{Version: "1.12.2"},
					},
					"1.12.3",
					true,
				),
				Entry("Should not find qualifying version - no higher version available that is not expired or in preview.",
					"1.12.2",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{
							Version:        "1.12.9",
							ExpirationDate: &expirationDateInThePast,
						},
						{
							Version:        "1.12.4",
							Classification: &previewClassification,
						},
						{Version: "1.12.2"},
					},
					"",
					false,
				),
				Entry("Should not find qualifying version - is already highest version of minor.",
					"1.12.2",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{Version: "1.12.2"},
					},
					"",
					false,
				),
				Entry("Should not find qualifying version - is already on latest version of latest minor.",
					"1.15.1",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{Version: "1.12.2"},
					},
					"",
					false,
				),
			)

			DescribeTable("#GetKubernetesVersionForMinorUpdate",
				func(currentVersion string, cloudProfileVersions []gardencorev1beta1.ExpirableVersion, expectedVersion string, qualifyingVersionFound bool) {
					cloudProfile := gardencorev1beta1.CloudProfile{
						Spec: gardencorev1beta1.CloudProfileSpec{
							Kubernetes: gardencorev1beta1.KubernetesSettings{
								Versions: cloudProfileVersions,
							},
						},
					}
					ok, newVersion, err := GetKubernetesVersionForMinorUpdate(&cloudProfile, currentVersion)
					Expect(err).ToNot(HaveOccurred())
					Expect(ok).To(Equal(qualifyingVersionFound))
					Expect(newVersion).To(Equal(expectedVersion))
				},
				Entry("Do not consider preview versions of the consecutive minor version.",
					"1.11.3",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{
							Version:        "1.12.9",
							Classification: &previewClassification,
						},
						{
							Version:        "1.12.4",
							Classification: &previewClassification,
						},
						// latest qualifying version for minor version update for version 1.11.3
						{Version: "1.12.3"},
						{Version: "1.12.2"},
						{Version: "1.11.3"},
					},
					"1.12.3",
					true,
				),
				Entry("Should find qualifying version - latest non-expired version of the consecutive minor version.",
					"1.11.3",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{
							Version:        "1.12.9",
							ExpirationDate: &expirationDateInThePast,
						},
						{
							Version:        "1.12.4",
							ExpirationDate: &expirationDateInThePast,
						},
						// latest qualifying version for updating version 1.11.3
						{Version: "1.12.3"},
						{Version: "1.12.2"},
						{Version: "1.11.3"},
						{Version: "1.10.1"},
						{Version: "1.09.0"},
					},
					"1.12.3",
					true,
				),
				// check that multiple consecutive minor versions are possible
				Entry("Should find qualifying version if there are only expired versions available in the consecutive minor version - pick latest expired version of that minor.",
					"1.11.3",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						// latest qualifying version for updating version 1.11.3
						{
							Version:        "1.12.9",
							ExpirationDate: &expirationDateInThePast,
						},
						{
							Version:        "1.12.4",
							ExpirationDate: &expirationDateInThePast,
						},
						{Version: "1.11.3"},
					},
					"1.12.9",
					true,
				),
				Entry("Should not find qualifying version - there is no consecutive minor version available.",
					"1.10.3",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{
							Version:        "1.12.9",
							ExpirationDate: &expirationDateInThePast,
						},
						{
							Version:        "1.12.4",
							ExpirationDate: &expirationDateInThePast,
						},
						{Version: "1.12.3"},
						{Version: "1.12.2"},
						{Version: "1.10.3"},
					},
					"",
					false,
				),
				Entry("Should not find qualifying version - already on latest minor version.",
					"1.15.1",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{Version: "1.12.2"},
					},
					"",
					false,
				),
				Entry("Should not find qualifying version - is already on latest version of latest minor version.",
					"1.15.1",
					[]gardencorev1beta1.ExpirableVersion{
						{Version: "1.15.1"},
						{Version: "1.15.0"},
						{Version: "1.14.4"},
						{Version: "1.12.2"},
					},
					"",
					false,
				),
			)
			DescribeTable("Test version filter predicates",
				func(predicate VersionPredicate, version *semver.Version, expirableVersion gardencorev1beta1.ExpirableVersion, expectFilterVersion, expectError bool) {
					shouldFilter, err := predicate(expirableVersion, version)
					if expectError {
						Expect(err).To(HaveOccurred())
						return
					}
					Expect(err).ToNot(HaveOccurred())
					Expect(shouldFilter).To(Equal(expectFilterVersion))
				},

				// #FilterDifferentMajorMinorVersion
				Entry("Should filter version - has not the same major.minor.",
					FilterDifferentMajorMinorVersion(*semver.MustParse("1.2.0")),
					semver.MustParse("1.1.1"),
					gardencorev1beta1.ExpirableVersion{},
					true,
					false,
				),
				Entry("Should filter version - version has same major.minor but is lower",
					FilterDifferentMajorMinorVersion(*semver.MustParse("1.1.2")),
					semver.MustParse("1.1.1"),
					gardencorev1beta1.ExpirableVersion{},
					true,
					false,
				),
				Entry("Should not filter version - has the same major.minor.",
					FilterDifferentMajorMinorVersion(*semver.MustParse("1.1.0")),
					semver.MustParse("1.1.1"),
					gardencorev1beta1.ExpirableVersion{},
					false,
					false,
				),

				// #FilterNonConsecutiveMinorVersion
				Entry("Should filter version - has not the consecutive minor version.",
					FilterNonConsecutiveMinorVersion(*semver.MustParse("1.3.0")),
					semver.MustParse("1.1.1"),
					gardencorev1beta1.ExpirableVersion{},
					true,
					false,
				),
				Entry("Should filter version - has the same minor version.",
					FilterNonConsecutiveMinorVersion(*semver.MustParse("1.1.0")),
					semver.MustParse("1.1.1"),
					gardencorev1beta1.ExpirableVersion{},
					true,
					false,
				),
				Entry("Should not filter version - has consecutive minor.",
					FilterNonConsecutiveMinorVersion(*semver.MustParse("1.1.0")),
					semver.MustParse("1.2.0"),
					gardencorev1beta1.ExpirableVersion{},
					false,
					false,
				),

				// #FilterSameVersion
				Entry("Should filter version.",
					FilterSameVersion(*semver.MustParse("1.1.1")),
					semver.MustParse("1.1.1"),
					gardencorev1beta1.ExpirableVersion{},
					true,
					false,
				),
				Entry("Should not filter version.",
					FilterSameVersion(*semver.MustParse("1.1.1")),
					semver.MustParse("1.1.2"),
					gardencorev1beta1.ExpirableVersion{},
					false,
					false,
				),

				// #FilterExpiredVersion
				Entry("Should filter expired version.",
					FilterExpiredVersion(),
					nil,
					gardencorev1beta1.ExpirableVersion{
						ExpirationDate: &expirationDateInThePast,
					},
					true,
					false,
				),
				Entry("Should not filter version - expiration date is not expired",
					FilterExpiredVersion(),
					nil,
					gardencorev1beta1.ExpirableVersion{
						ExpirationDate: &metav1.Time{Time: time.Now().Add(time.Hour)},
					},
					false,
					false,
				),
				Entry("Should not filter version.",
					FilterExpiredVersion(),
					nil,
					gardencorev1beta1.ExpirableVersion{},
					false,
					false,
				),
				// #FilterLowerVersion
				Entry("Should filter version - version is lower",
					FilterLowerVersion(*semver.MustParse("1.1.1")),
					semver.MustParse("1.1.0"),
					gardencorev1beta1.ExpirableVersion{},
					true,
					false,
				),
				Entry("Should not filter version - version is higher / equal",
					FilterLowerVersion(*semver.MustParse("1.1.1")),
					semver.MustParse("1.1.2"),
					gardencorev1beta1.ExpirableVersion{},
					false,
					false,
				),
			)
		})
	})
})