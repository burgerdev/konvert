package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"sigs.k8s.io/yaml"
)

var (
	from = flag.String("from", "pod", "input kind")
	to   = flag.String("to", "deployment", "output kind")
)

func parse(r io.Reader) (*corev1.Pod, error) {
	// TODO(burgerdev): allow other sources
	switch *from {
	case "pod":
		return parseAs[corev1.Pod](r)
	}
	return nil, fmt.Errorf("no conversion from %q to v1.Pod", *from)
}

func parseAs[A any](r io.Reader) (*A, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("reading input: %w", err)
	}
	a := new(A)
	if err := yaml.Unmarshal(data, a); err != nil {
		return nil, fmt.Errorf("unmarshaling v1.Pod: %w", err)
	}
	return a, nil
}

func convert(pod *corev1.Pod) (any, error) {
	// TODO(burgerdev): allow other sinks
	switch *to {
	case "deployment":
		return convertToDeployment(pod)
	}
	return nil, fmt.Errorf("no conversion to %q from v1.Pod", *to)
}

func convertToDeployment(pod *corev1.Pod) (*appsv1.Deployment, error) {
	d := &appsv1.Deployment{}
	d.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "Deployment",
	})

	d.SetName(pod.Name)
	d.SetNamespace(pod.Namespace)
	d.SetLabels(pod.Labels)
	d.SetAnnotations(pod.Annotations)
	d.Spec.Selector = &metav1.LabelSelector{MatchLabels: pod.Labels}
	d.Spec.Template.Labels = pod.Labels
	d.Spec.Template.Annotations = pod.Annotations
	d.Spec.Template.Spec = pod.Spec

	return d, nil
}

func main() {
	flag.Parse()

	pod, err := parse(os.Stdin)
	must(err)
	obj, err := convert(pod)
	must(err)
	data, err := yaml.Marshal(obj)
	must(err)
	os.Stdout.Write(data)
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
