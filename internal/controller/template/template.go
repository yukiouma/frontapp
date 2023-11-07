package template

import (
	"bytes"

	"text/template"

	frontierv1 "github.com/yukiouma/frontapp/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func parse(name, templateText string, app *frontierv1.FrontApp) ([]byte, error) {
	tmpl, err := template.New(name).Parse(templateText)
	if err != nil {
		return nil, err
	}
	b := bytes.NewBuffer([]byte{})
	err = tmpl.Execute(b, app)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func NewConfig(app *frontierv1.FrontApp) (*corev1.ConfigMap, error) {
	result := corev1.ConfigMap{}
	tmpl, err := parse("configmap", CONFIG_MAP_TEMPLATE, app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(tmpl, &result)
	return &result, err
}

func NewDeployment(app *frontierv1.FrontApp) (*appsv1.Deployment, error) {
	result := appsv1.Deployment{}
	tmpl, err := parse("deployment", DEPLOYMENT_TEMPLATE, app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(tmpl, &result)
	return &result, err
}

func NewService(app *frontierv1.FrontApp) (*corev1.Service, error) {
	result := corev1.Service{}
	tmpl, err := parse("service", SERVICE_TEMPLATE, app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(tmpl, &result)
	return &result, err
}

func NewIngress(app *frontierv1.FrontApp) (*netv1.Ingress, error) {
	result := netv1.Ingress{}
	tmpl, err := parse("ingress", INGRESS_TEMPLATE, app)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(tmpl, &result)
	return &result, err
}

const (
	CONFIG_MAP_TEMPLATE = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{.ObjectMeta.Name}}
  namespace: {{.ObjectMeta.Namespace}}
data:
  Caddyfile: |
    :80 {
      root * /usr/share/caddy
      file_server
      reverse_proxy /api/* {{ .Spec.ReverseProxy }} {
          header_up Host {http.reverse_proxy.upstream.hostport}
          header_down Access-Control-Allow-Headers *
          header_down Access-Control-Allow-Origin *
      }
    }
`

	SERVICE_TEMPLATE = `
apiVersion: v1
kind: Service
metadata:
  name: {{.ObjectMeta.Name}}
  namespace: {{.ObjectMeta.Namespace}}
spec:
  ports:
    - name: {{.ObjectMeta.Name}}
      port: 80
      protocol: TCP
      targetPort: 80
  selector:
    app: {{.ObjectMeta.Name}}
  sessionAffinity: None
  type: ClusterIP
`

	DEPLOYMENT_TEMPLATE = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.ObjectMeta.Name}}
  namespace: {{.ObjectMeta.Namespace}}
spec:
  replicas: 1
  selector:
    matchLabels:
      app: {{.ObjectMeta.Name}}
  template:
    metadata:
      labels:
        app: {{.ObjectMeta.Name}}
    spec:
      containers:
        - name: {{.ObjectMeta.Name}}
          image: {{.Spec.Image}}
          ports:
            - containerPort: 80
              protocol: TCP
          volumeMounts:
          - name: caddy
            mountPath: "/etc/caddy"
            readOnly: true
          resources:
            requests:
              memory: "128Mi"
              cpu: "100m"
            limits:
              memory: "1Gi"
              cpu: "1"
      volumes:
      - name: caddy
        configMap:
          name: {{.ObjectMeta.Name}}
  
`

	INGRESS_TEMPLATE = `
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: {{.ObjectMeta.Name}}
  namespace: {{.ObjectMeta.Namespace}}
spec:
  rules:
    - host: {{ .Spec.Url }}
      http:
        paths:
          - path: /
            backend:
              service:
                name: {{.ObjectMeta.Name}}
                port:
                  number: 80
            pathType: ImplementationSpecific
`
)
