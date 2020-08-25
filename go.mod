module github.com/fairwindsops/polaris

go 1.13

require (
	cloud.google.com/go v0.64.0
	contrib.go.opencensus.io/exporter/ocagent v0.7.0
	git.apache.org/thrift.git v0.12.0 // indirect
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.4 // indirect
	github.com/Azure/go-autorest/autorest/adal v0.9.2 // indirect
	github.com/appscode/jsonpatch v1.0.1
	github.com/beorn7/perks v1.0.1
	github.com/bombsimon/logrusr v0.0.0-20200131103305-03a291ce59b4
	github.com/census-instrumentation/opencensus-proto v0.2.1
	github.com/coreos/go-oidc v2.1.0+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-logr/logr v0.1.0
	github.com/go-logr/zapr v0.1.0
	github.com/go-openapi/validate v0.19.5 // indirect
	github.com/gobuffalo/depgen v0.1.0 // indirect
	github.com/gobuffalo/envy v1.9.0
	github.com/gobuffalo/genny v0.6.0
	github.com/gobuffalo/gogen v0.2.0
	github.com/gobuffalo/logger v1.0.3
	github.com/gobuffalo/mapi v1.2.1
	github.com/gobuffalo/packd v1.0.0
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gobuffalo/syncx v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/groupcache v0.0.0-20200121045136-8c9f03a8e57e
	github.com/golang/lint v0.0.0-20180702182130-06c8688daad7 // indirect
	github.com/golang/protobuf v1.4.2
	github.com/google/btree v1.0.0
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/gofuzz v1.2.0
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1
	github.com/gophercloud/gophercloud v0.12.0
	github.com/gorilla/mux v1.8.0
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79
	github.com/grpc-ecosystem/grpc-gateway v1.14.6
	github.com/hashicorp/golang-lru v0.5.4
	github.com/imdario/mergo v0.3.11
	github.com/joho/godotenv v1.3.0
	github.com/json-iterator/go v1.1.10
	github.com/karrick/godirwalk v1.16.1
	github.com/konsorten/go-windows-terminal-sequences v1.0.3
	github.com/kr/pretty v0.2.0 // indirect
	github.com/markbates/oncer v1.0.0
	github.com/markbates/safe v1.0.1
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/onsi/gomega v1.10.1 // indirect
	github.com/pborman/uuid v1.2.0
	github.com/petar/GoLLRB v0.0.0-20190514000832-33fb24c13b99
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/pmezard/go-difflib v1.0.0
	github.com/pquerna/cachecontrol v0.0.0-20171018203845-0dec1b30a021 // indirect
	github.com/prometheus/client_golang v1.7.1
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.13.0
	github.com/prometheus/procfs v0.1.3
	github.com/qri-io/jsonpointer v0.1.1
	github.com/qri-io/jsonschema v0.1.1
	github.com/rogpeppe/go-internal v1.6.0
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/stoewer/go-strcase v1.2.0 // indirect
	github.com/stretchr/testify v1.6.1
	gitlab.com/golang-commonmark/html v0.0.0-20191124015941-a22733972181
	gitlab.com/golang-commonmark/linkify v0.0.0-20200225224916-64bca66f6ad3
	gitlab.com/golang-commonmark/markdown v0.0.0-20191127184510-91b5b3c99c19
	gitlab.com/golang-commonmark/mdurl v0.0.0-20191124015652-932350d1cb84
	gitlab.com/golang-commonmark/puny v0.0.0-20191124015043-9f83538fa04f
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738 // indirect
	go.opencensus.io v0.22.4
	go.uber.org/atomic v1.6.0
	go.uber.org/multierr v1.5.0
	go.uber.org/zap v1.15.0
	golang.org/x/crypto v0.0.0-20200820211705-5c72a883971a
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	golang.org/x/sys v0.0.0-20200824131525-c12d262b63d8
	golang.org/x/text v0.3.3
	golang.org/x/time v0.0.0-20200630173020-3af7569d3a1e
	golang.org/x/tools v0.0.0-20200817023811-d00afeaade8f
	gomodules.xyz/jsonpatch/v2 v2.1.0 // indirect
	gonum.org/v1/netlib v0.0.0-20190331212654-76723241ea4e // indirect
	google.golang.org/api v0.30.0
	google.golang.org/appengine v1.6.6
	google.golang.org/genproto v0.0.0-20200815001618-f69a88009b70
	google.golang.org/grpc v1.31.0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/square/go-jose.v2 v2.2.2 // indirect
	gopkg.in/yaml.v2 v2.3.0
	gopkg.in/yaml.v3 v3.0.0-20200615113413-eeeca48fe776
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.4
	k8s.io/code-generator v0.18.6 // indirect
	k8s.io/gengo v0.0.0-20200413195148-3a45101e95ac // indirect
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.1.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/utils v0.0.0-20200821003339-5e75c0163111 // indirect
	sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.7 // indirect
	sigs.k8s.io/controller-runtime v0.6.1
	sigs.k8s.io/structured-merge-diff v0.0.0-20190525122527-15d366b2352e // indirect
	sigs.k8s.io/structured-merge-diff/v2 v2.0.1 // indirect
	sigs.k8s.io/yaml v1.2.0
)
