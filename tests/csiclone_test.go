package tests

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/types"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"kubevirt.io/containerized-data-importer/pkg/controller"
	"kubevirt.io/containerized-data-importer/tests/framework"
	"kubevirt.io/containerized-data-importer/tests/utils"

	cdiv1 "kubevirt.io/containerized-data-importer/pkg/apis/core/v1beta1"
)

var _ = Describe("[vendor:cnv-qe@redhat.com][level:component][crit:high] CSI Volume cloning tests", func() {
	var originalProfileSpec *cdiv1.StorageProfileSpec

	fillData := "123456789012345678901234567890123456789012345678901234567890"
	testFile := utils.DefaultPvcMountPath + "/source.txt"
	fillCommandFilesystem := "echo -n \"" + fillData + "\" >> " + testFile

	f := framework.NewFramework("dv-func-test")
	getStorageProfileSpec := func(storageClassName string) *cdiv1.StorageProfileSpec {
		storageProfile, err := f.CdiClient.CdiV1beta1().StorageProfiles().Get(context.TODO(), storageClassName, metav1.GetOptions{})
		Expect(err).ToNot(HaveOccurred())
		By(fmt.Sprintf("inside Got original storage profile.spec: %v", storageProfile.Spec))

		originalProfileSpec := storageProfile.Spec.DeepCopy()
		return originalProfileSpec
	}

	updateStorageProfileSpec := func(client client.Client, name string, spec cdiv1.StorageProfileSpec) {
		storageProfile := &cdiv1.StorageProfile{}
		err := client.Get(context.TODO(), types.NamespacedName{Name: name}, storageProfile)
		Expect(err).ToNot(HaveOccurred())
		storageProfile.Spec = spec
		err = client.Update(context.TODO(), storageProfile)
		Expect(err).ToNot(HaveOccurred())
	}

	configureCloneStrategy := func(client client.Client,
		storageClassName string,
		spec *cdiv1.StorageProfileSpec) {

		cloneStrategyCsiClone := cdiv1.CDICloneStrategy(cdiv1.CloneStrategyCsiClone)
		newProfileSpec := updateCloneStrategy(spec, cloneStrategyCsiClone)
		updateStorageProfileSpec(client, storageClassName, *newProfileSpec)

		Eventually(func() *cdiv1.CDICloneStrategy {
			profile, err := f.CdiClient.CdiV1beta1().StorageProfiles().Get(context.TODO(), storageClassName, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			if len(profile.Status.ClaimPropertySets) > 0 {
				return profile.Status.ClaimPropertySets[0].CloneStrategy
			}
			return nil
		}, time.Second*30, time.Second).Should(Equal(&cloneStrategyCsiClone))
	}

	BeforeEach(func() {
		cloneStorageClassName := f.CsiCloneSCName
		By(fmt.Sprintf("Get original storage profile: %s", cloneStorageClassName))

		originalProfileSpec = getStorageProfileSpec(cloneStorageClassName)
		By(fmt.Sprintf("Got original storage profile: %v", originalProfileSpec))

	})

	AfterEach(func() {
		cloneStorageClassName := f.CsiCloneSCName
		By("Restore the profile")
		updateStorageProfileSpec(f.CrClient, cloneStorageClassName, *originalProfileSpec)
	})

	It("Verify DataVolume CSI Volume Cloning - volumeMode filesystem - Positive flow", func() {
		if !f.IsCSIVolumeCloneStorageClassAvailable() {
			Skip("CSI Volume Clone is not applicable")
		}

		By(fmt.Sprintf("configure storage profile %s", f.CsiCloneSCName))
		configureCloneStrategy(f.CrClient, f.CsiCloneSCName, originalProfileSpec)

		dataVolume, md5 := createDataVolume("dv-csi-clone-test-1", fillCommandFilesystem, v1.PersistentVolumeFilesystem, f.CsiCloneSCName, f)
		verifyEvent(string(cdiv1.CSICloneInProgress), dataVolume.Namespace, f)
		// Wait for operation Succeeded
		waitForDvPhase(cdiv1.Succeeded, dataVolume, f)
		verifyEvent(controller.CloneSucceeded, dataVolume.Namespace, f)
		// Verify PVC's content
		verifyPVC(dataVolume, f, testFile, md5)
	})

	It("Verify DataVolume Smart Cloning - volumeMode block - Positive flow", func() {
		if !f.IsCSIVolumeCloneStorageClassAvailable() {
			Skip("CSI Volume Clone is not applicable")
		}

		By(fmt.Sprintf("configure storage profile %s", f.CsiCloneSCName))
		configureCloneStrategy(f.CrClient, f.CsiCloneSCName, originalProfileSpec)

		dataVolume, expectedMd5 := createDataVolume("dv-csi-clone-test-1", utils.DefaultPvcMountPath, v1.PersistentVolumeBlock, f.CsiCloneSCName, f)
		verifyEvent(controller.CSICloneInProgress, dataVolume.Namespace, f)
		// Wait for operation Succeeded
		waitForDvPhase(cdiv1.Succeeded, dataVolume, f)
		verifyEvent(controller.CloneSucceeded, dataVolume.Namespace, f)
		// Verify PVC's content
		verifyPVC(dataVolume, f, utils.DefaultPvcMountPath, expectedMd5)
	})

	It("Verify DataVolume Smart Cloning - volumeMode filesystem - Waits for source to be available", func() {
		if !f.IsSnapshotStorageClassAvailable() {
			Skip("Smart Clone is not applicable")
		}

		By(fmt.Sprintf("configure storage profile %s", f.CsiCloneSCName))
		configureCloneStrategy(f.CrClient, f.CsiCloneSCName, originalProfileSpec)
		sourcePvc := createAndPopulateSourcePVC("dv-smart-clone-test-1", v1.PersistentVolumeFilesystem, f.CsiCloneSCName, f)
		pod, err := f.CreateExecutorPodWithPVC("temp-pod", f.Namespace.Name, sourcePvc)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() bool {
			pod, err = f.K8sClient.CoreV1().Pods(f.Namespace.Name).Get(context.TODO(), pod.Name, metav1.GetOptions{})
			Expect(err).ToNot(HaveOccurred())
			return pod.Status.Phase == v1.PodRunning
		}, 90*time.Second, 2*time.Second).Should(BeTrue())

		By(fmt.Sprintf("creating a new target PVC (datavolume) to clone %s", sourcePvc.Name))
		dataVolume := utils.NewCloningDataVolume(dataVolumeName, "1Gi", sourcePvc)
		if f.CsiCloneSCName != "" {
			dataVolume.Spec.PVC.StorageClassName = &f.CsiCloneSCName
		}

		By(fmt.Sprintf("creating new datavolume %s", dataVolume.Name))
		dataVolume, err = utils.CreateDataVolumeFromDefinition(f.CdiClient, f.Namespace.Name, dataVolume)
		Expect(err).ToNot(HaveOccurred())

		verifyEvent(controller.CSICloneSourceInUse, dataVolume.Namespace, f)
		err = f.K8sClient.CoreV1().Pods(f.Namespace.Name).Delete(context.TODO(), pod.Name, metav1.DeleteOptions{})
		Expect(err).ToNot(HaveOccurred())

		verifyEvent(controller.CSICloneInProgress, dataVolume.Namespace, f)
		// Wait for operation Succeeded
		waitForDvPhase(cdiv1.Succeeded, dataVolume, f)
		verifyEvent(controller.CloneSucceeded, dataVolume.Namespace, f)
		// Verify PVC's content
		verifyPVC(dataVolume, f, utils.DefaultPvcMountPath, utils.TinyCoreBlockMD5)
	})
})

func updateCloneStrategy(originalProfileSpec *cdiv1.StorageProfileSpec, cloneStrategy cdiv1.CDICloneStrategy) *cdiv1.StorageProfileSpec {
	newProfileSpec := originalProfileSpec.DeepCopy()

	if len(newProfileSpec.ClaimPropertySets) == 0 {
		newProfileSpec.ClaimPropertySets = []cdiv1.ClaimPropertySet{{CloneStrategy: &cloneStrategy}}
	} else {
		newProfileSpec.ClaimPropertySets[0].CloneStrategy = &cloneStrategy
	}

	return newProfileSpec
}
