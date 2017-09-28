package main

import (
  "fmt"
  "flag"
  "time"
  "sync"
  "context"

  "k8s.io/client-go/rest"
  "k8s.io/client-go/pkg/api/v1"
  "k8s.io/client-go/tools/cache"
  "k8s.io/apimachinery/pkg/fields"
  "k8s.io/apimachinery/pkg/runtime"
  "k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
  metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type APIFooList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []APIFoo `json:"items"`
}

type APIFoo struct {
	metav1.TypeMeta    `json:",inline"`
	metav1.ObjectMeta  `json:"metadata,omitempty"`
	Spec               APIFooSpec `json:"spec"`
}

type APIFooSpec struct {
  Foo string `json:"foo"`
}

type Controller struct {
  RESTClient *rest.RESTClient
}

var (
	schemeBuilder      = runtime.NewSchemeBuilder(addKnownTypes)
	addToScheme        = schemeBuilder.AddToScheme
	schemeGroupVersion = schema.GroupVersion{Group: "bar.io", Version: "v1"}
)

func main() {
  kubeconfig := flag.String("kubeconfig", "", "")
	flag.Parse()
  ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
  restClient, err := getRESTClient(*kubeconfig)
  if err != nil {
    panic(err)
  }
  controller := Controller{restClient}

	go controller.run(ctx)

  wg := sync.WaitGroup{}
  wg.Add(1)
  wg.Wait()
}

func (c Controller) run(ctx context.Context) error {
  if err := c.watch(ctx); err != nil {
    return err
  }
	<-ctx.Done()
	return ctx.Err()
}

func (c Controller) watch(ctx context.Context) error {
  source := cache.NewListWatchFromClient(
		c.RESTClient,
		"apifoos",
		v1.NamespaceAll,
		fields.Everything(),
	)
	_, ctlr := cache.NewInformer(
		source,
		&APIFoo{},
		5*time.Minute,
		cache.ResourceEventHandlerFuncs{
      AddFunc: func(obj interface{}) {
        item, _ := obj.(*APIFoo)
        fmt.Printf("adding crd named %s", item.ObjectMeta.Name)
      },
      UpdateFunc: func(old, new interface{}) {
        item, _ := old.(*APIFoo)
        fmt.Printf("updating crd named %s", item.ObjectMeta.Name)
      },
      DeleteFunc: func(obj interface{}) {
        item, _ := obj.(*APIFoo)
        fmt.Printf("deleting crd named %s", item.ObjectMeta.Name)
      },
    },
	)
  go ctlr.Run(ctx.Done())
	return nil
}

func getRESTClient(path string) (*rest.RESTClient, error) {
  cfg, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	if err := addToScheme(scheme); err != nil {
		return nil, err
	}
	config := *cfg
	config.APIPath = "/apis"
	config.GroupVersion = &schemeGroupVersion
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}
	restClient, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}
  return restClient, nil
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(schemeGroupVersion,
		&APIFoo{},
		&APIFooList{},
	)
	metav1.AddToGroupVersion(scheme, schemeGroupVersion)
	return nil
}