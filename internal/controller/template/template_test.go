package template

import (
	"strings"
	"testing"

	frontierv1 "github.com/yukiouma/frontapp/api/v1"
	corev1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	app = frontierv1.FrontApp{
		ObjectMeta: corev1.ObjectMeta{
			Name:      "demo",
			Namespace: "default",
		},
		Spec: frontierv1.FrontAppSpec{
			Image:        "nginx:apline",
			ReverseProxy: "www.example.com",
			Url:          "www.demo.com",
		},
	}
)

func TestNewConfigMap(t *testing.T) {
	obj, err := NewConfig(&app)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(obj.Data["Caddyfile"], app.Spec.ReverseProxy) {
		t.Fail()
	}
}

func TestNewDeployment(t *testing.T) {
	obj, err := NewDeployment(&app)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Spec.Template.Spec.Containers[0].Image != app.Spec.Image {
		t.Fail()
	}
}

func TestNewService(t *testing.T) {
	obj, err := NewService(&app)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Spec.Ports[0].TargetPort.IntValue() != 80 {
		t.Fail()
	}
}

func TestNewIngress(t *testing.T) {
	obj, err := NewIngress(&app)
	if err != nil {
		t.Fatal(err)
	}
	if obj.Spec.Rules[0].Host != app.Spec.Url {
		t.Fail()
	}
}
