package webhook

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-app-mesh-inject/pkg/config"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/fake"
)

const admissionReview = `
{
  "kind": "AdmissionReview",
  "apiVersion": "admission.k8s.io/v1beta1",
  "request": {
    "uid": "53ad2101-497a-11e9-960e-0edb66a862f2",
    "kind": {
      "group": "",
      "version": "v1",
      "kind": "Pod"
    },
    "resource": {
      "group": "",
      "version": "v1",
      "resource": "pods"
    },
    "namespace": "test",
    "operation": "CREATE",
    "userInfo": {
      "username": "system:unsecured",
      "groups": [
        "system:masters",
        "system:authenticated"
      ]
    },
    "object": {
      "metadata": {
        "generateName": "podinfo-7c45b75c87-",
        "creationTimestamp": null,
        "labels": {
          "app": "podinfo",
          "pod-template-hash": "3701631743"
        },
        "annotations": {
          "appmesh.k8s.aws/ports": "9898",
          "appmesh.k8s.aws/egress_ignored_ports": "22",
          "appmesh.k8s.aws/virtualNode": "podinfo"
        }
      },
      "spec": {
        "volumes": [
          {
            "name": "default-token-xhfkr",
            "secret": {
              "secretName": "default-token-xhfkr"
            }
          }
        ],
        "containers": [
          {
            "name": "podinfod",
            "image": "quay.io/stefanprodan/podinfo:1.4.0",
            "command": [
              "./podinfo",
              "--port=9898"
            ],
            "ports": [
              {
                "name": "http",
                "containerPort": 9898,
                "protocol": "TCP"
              }
            ],
            "volumeMounts": [
              {
                "name": "default-token-xhfkr",
                "readOnly": true,
                "mountPath": "/var/run/secrets/kubernetes.io/serviceaccount"
              }
            ],
            "terminationMessagePath": "/dev/termination-log",
            "terminationMessagePolicy": "File",
            "imagePullPolicy": "IfNotPresent"
          }
        ],
        "restartPolicy": "Always",
        "terminationGracePeriodSeconds": 30,
        "dnsPolicy": "ClusterFirst",
        "serviceAccountName": "default",
        "serviceAccount": "default",
        "securityContext": {},
        "schedulerName": "default-scheduler",
        "tolerations": [
          {
            "key": "node.kubernetes.io/not-ready",
            "operator": "Exists",
            "effect": "NoExecute",
            "tolerationSeconds": 300
          },
          {
            "key": "node.kubernetes.io/unreachable",
            "operator": "Exists",
            "effect": "NoExecute",
            "tolerationSeconds": 300
          }
        ],
        "priority": 0
      },
      "status": {}
    },
    "oldObject": null
  }
}
`

func mockServer() *Server {
	cfg := config.Config{
		Port:     8080,
		MeshName: "global",
		Region:   "us-east-1",
		LogLevel: "debug",
	}

	scheme := runtime.NewScheme()
	corev1.AddToScheme(scheme)
	admissionregistrationv1beta1.AddToScheme(scheme)
	codecs := serializer.NewCodecFactory(scheme)
	kubeDecoder := codecs.UniversalDeserializer()

	namespace := &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{APIVersion: corev1.SchemeGroupVersion.String()},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Labels: map[string]string{
				sidecarInjectAnnotation: "enabled",
			},
		},
	}
	kubeClient := fake.NewSimpleClientset(namespace)

	return &Server{
		Config:      cfg,
		KubeClient:  kubeClient,
		KubeDecoder: kubeDecoder,
	}
}

func TestServer_Inject(t *testing.T) {
	srv := mockServer()

	req, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte(admissionReview)))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.injectHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "\"allowed\":true") {
		t.Errorf("handler returned wrong result")
	}
}

func TestServer_Health(t *testing.T) {
	srv := mockServer()

	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(srv.healthHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
