/*
Copyright 2026.

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

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	pgv1alpha1 "github.com/Kitio-Tek/vigil-kubernetes/api/v1alpha1"
	"github.com/Kitio-Tek/vigil-kubernetes/internal/pgbouncer"
)

var _ = Describe("PostgresPooler Controller", func() {
	const (
		poolerCluster = "pooler-test-cluster"
		poolerNS      = "default"
		timeout       = time.Second * 10
		interval      = time.Millisecond * 250
	)

	ctx := context.Background()
	clusterKey := types.NamespacedName{Name: poolerCluster, Namespace: poolerNS}

	newCluster := func(annotations map[string]string) *pgv1alpha1.PostgresCluster {
		c := &pgv1alpha1.PostgresCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:        poolerCluster,
				Namespace:   poolerNS,
				Annotations: annotations,
			},
			Spec: pgv1alpha1.PostgresClusterSpec{
				PostgresVersion: 16,
				Instances:       1,
				Storage: pgv1alpha1.StorageSpec{
					Size: resource.MustParse("1Gi"),
				},
			},
		}
		return c
	}

	poolerReconciler := func() *PostgresPoolerReconciler {
		return &PostgresPoolerReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	}

	reconcileOnce := func() (reconcile.Result, error) {
		return poolerReconciler().Reconcile(ctx, reconcile.Request{NamespacedName: clusterKey})
	}

	AfterEach(func() {
		cluster := &pgv1alpha1.PostgresCluster{}
		if err := k8sClient.Get(ctx, clusterKey, cluster); err == nil {
			cluster.Finalizers = nil
			_ = k8sClient.Update(ctx, cluster)
			_ = k8sClient.Delete(ctx, cluster)
		}
		// Clean up pooler resources.
		cm := &corev1.ConfigMap{}
		cmKey := types.NamespacedName{Name: pgbouncer.ConfigMapName(poolerCluster), Namespace: poolerNS}
		if err := k8sClient.Get(ctx, cmKey, cm); err == nil {
			_ = k8sClient.Delete(ctx, cm)
		}
		deploy := &appsv1.Deployment{}
		depKey := types.NamespacedName{Name: pgbouncer.DeploymentName(poolerCluster), Namespace: poolerNS}
		if err := k8sClient.Get(ctx, depKey, deploy); err == nil {
			_ = k8sClient.Delete(ctx, deploy)
		}
		svc := &corev1.Service{}
		svcKey := types.NamespacedName{Name: pgbouncer.ServiceName(poolerCluster), Namespace: poolerNS}
		if err := k8sClient.Get(ctx, svcKey, svc); err == nil {
			_ = k8sClient.Delete(ctx, svc)
		}
	})

	Describe("Cluster not found", func() {
		It("should return no error when the cluster does not exist", func() {
			_, err := reconcileOnce()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Pooler annotation absent", func() {
		BeforeEach(func() {
			Expect(k8sClient.Create(ctx, newCluster(nil))).To(Succeed())
		})

		It("should not create any pooler resources when the annotation is missing", func() {
			_, err := reconcileOnce()
			Expect(err).NotTo(HaveOccurred())

			cm := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      pgbouncer.ConfigMapName(poolerCluster),
				Namespace: poolerNS,
			}, cm)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})

	Describe("Pooler annotation present", func() {
		BeforeEach(func() {
			c := newCluster(map[string]string{
				"pg.vigil.io/enable-pooler": "true",
			})
			Expect(k8sClient.Create(ctx, c)).To(Succeed())
		})

		It("should create a ConfigMap for pgbouncer.ini", func() {
			_, err := reconcileOnce()
			Expect(err).NotTo(HaveOccurred())

			cm := &corev1.ConfigMap{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      pgbouncer.ConfigMapName(poolerCluster),
				Namespace: poolerNS,
			}, cm)).To(Succeed())
			Expect(cm.Data).To(HaveKey("pgbouncer.ini"))
		})

		It("should create a Deployment with the pgbouncer container", func() {
			_, err := reconcileOnce()
			Expect(err).NotTo(HaveOccurred())

			deploy := &appsv1.Deployment{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      pgbouncer.DeploymentName(poolerCluster),
				Namespace: poolerNS,
			}, deploy)).To(Succeed())
			Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(deploy.Spec.Template.Spec.Containers[0].Name).To(Equal("pgbouncer"))
		})

		It("should create a Service targeting the pooler pods", func() {
			_, err := reconcileOnce()
			Expect(err).NotTo(HaveOccurred())

			svc := &corev1.Service{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      pgbouncer.ServiceName(poolerCluster),
				Namespace: poolerNS,
			}, svc)).To(Succeed())
			Expect(svc.Spec.Selector).To(HaveKeyWithValue("pg.vigil.io/cluster", poolerCluster))
		})

		It("should reconcile idempotently without error", func() {
			for i := 0; i < 3; i++ {
				_, err := reconcileOnce()
				Expect(err).NotTo(HaveOccurred())
			}
		})
	})

	Describe("Cluster paused", func() {
		BeforeEach(func() {
			c := newCluster(map[string]string{
				"pg.vigil.io/enable-pooler": "true",
			})
			c.Spec.Paused = true
			Expect(k8sClient.Create(ctx, c)).To(Succeed())
		})

		It("should not create pooler resources when the cluster is paused", func() {
			_, err := reconcileOnce()
			Expect(err).NotTo(HaveOccurred())

			cm := &corev1.ConfigMap{}
			err = k8sClient.Get(ctx, types.NamespacedName{
				Name:      pgbouncer.ConfigMapName(poolerCluster),
				Namespace: poolerNS,
			}, cm)
			Expect(errors.IsNotFound(err)).To(BeTrue())
		})
	})
})
