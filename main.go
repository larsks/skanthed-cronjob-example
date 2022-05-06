package main

import (
	"context"
	"flag"
	"log"
	"path/filepath"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type (
	ReportSpec struct {
		DatabaseHostName string
		DatabasePort     string
	}

	Report struct {
		Name string
		Spec ReportSpec
	}
)

func newCronJobFromReport(d *Report) *batchv1.CronJob {
	cronjob := &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			Kind: "Cronjob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: d.Name,
		},
		Spec: batchv1.CronJobSpec{
			ConcurrencyPolicy: "Forbid",
			Schedule:          "*/5 * * * *",
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "dayreport",
									Image: "docker.io/postgres:14",

									// Use environment variable names documented in
									// https://www.postgresql.org/docs/current/libpq-envars.html
									Env: []corev1.EnvVar{
										{
											Name: "PGDATABASE",

											// Get the database name from the POSTGRES_DB key in the "pg-config" Secret.
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "pg-config",
													},
													Key: "POSTGRES_DB",
												},
											},
										},
										{
											Name: "PGUSER",

											// Get the database user from the POSTGRES_USER key in the "pg-config" Secret.
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "pg-config",
													},
													Key: "POSTGRES_USER",
												},
											},
										},
										{
											Name: "PGPASSWORD",

											// Get the postgres password from the POSTGRES_PASSWORD key in the "pg-config" Secret.
											ValueFrom: &corev1.EnvVarSource{
												SecretKeyRef: &corev1.SecretKeySelector{
													LocalObjectReference: corev1.LocalObjectReference{
														Name: "pg-config",
													},
													Key: "POSTGRES_PASSWORD",
												},
											},
										},
										{
											Name:  "PGHOST",
											Value: d.Spec.DatabaseHostName,
										},
										{
											Name:  "PGPORT",
											Value: d.Spec.DatabasePort,
										},
									},

									// If you were trying to run psql using the shell, you would write it
									// like this (because, as we discussed, the shell script must be provided
									// as the value of the "-c" argument):
									//
									//    Command: []string{"sh", "-c", "psql -c \"SELECT current_timestamp;\""},
									//
									// But there's no reason to use the shell here, since we don't need any sort
									// of variable expansion or other shell features, so we can exec
									// psql directly:

									Command: []string{"psql", "-c", "SELECT current_timestamp;"},
								},
							},

							RestartPolicy: "Never",
						},
					},
				},
			},
		},
	}
	return cronjob

}

func main() {
	report := Report{
		Name: "example",
		Spec: ReportSpec{
			DatabaseHostName: "postgres",
			DatabasePort:     "5432",
		},
	}

	cronJob := newCronJobFromReport(&report)

	var default_kubeconfig string

	if home := homedir.HomeDir(); home != "" {
		default_kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		default_kubeconfig = ""
	}

	var namespace string
	var kubeconfig string

	// Accept either "-k <kubeconfig>" or "-kubeconfig <kubeconfig>"
	flag.StringVar(&kubeconfig, "kubeconfig", default_kubeconfig, "absolute path to the kubeconfig file")
	flag.StringVar(&kubeconfig, "k", default_kubeconfig, "absolute path to the kubeconfig file")

	// Accept either "-n <namespace>" or "-namespace <namespace>"
	flag.StringVar(&namespace, "namespace", "default", "namespace in which to create cronjob")
	flag.StringVar(&namespace, "n", "default", "namespace in which to create cronjob")
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	cronjobs := clientset.BatchV1().CronJobs(namespace)
	log.Printf("creating CronJob in namespace %s", namespace)
	if _, err = cronjobs.Create(context.TODO(), cronJob, metav1.CreateOptions{}); err != nil {
		panic(err.Error())
	}
}
