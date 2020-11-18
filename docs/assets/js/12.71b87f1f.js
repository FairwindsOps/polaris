(window.webpackJsonp=window.webpackJsonp||[]).push([[12],{370:function(e,t,r){"use strict";r.r(t);var s=r(42),o=Object(s.a)({},(function(){var e=this,t=e.$createElement,r=e._self._c||t;return r("ContentSlotsDistributor",{attrs:{"slot-key":e.$parent.slotKey}},[r("h1",{attrs:{id:"efficiency"}},[r("a",{staticClass:"header-anchor",attrs:{href:"#efficiency"}},[e._v("#")]),e._v(" Efficiency")]),e._v(" "),r("p",[e._v("Polaris supports a number of checks related to CPU and Memory requests and limits.")]),e._v(" "),r("h2",{attrs:{id:"presence-checks"}},[r("a",{staticClass:"header-anchor",attrs:{href:"#presence-checks"}},[e._v("#")]),e._v(" Presence Checks")]),e._v(" "),r("p",[e._v("To simplify ensure that these values have been set, the following attributes are available:")]),e._v(" "),r("table",[r("thead",[r("tr",[r("th",[e._v("key")]),e._v(" "),r("th",[e._v("default")]),e._v(" "),r("th",[e._v("description")])])]),e._v(" "),r("tbody",[r("tr",[r("td",[r("code",[e._v("resources.cpuRequestsMissing")])]),e._v(" "),r("td",[r("code",[e._v("warning")])]),e._v(" "),r("td",[e._v("Fails when "),r("code",[e._v("resources.requests.cpu")]),e._v(" attribute is not configured.")])]),e._v(" "),r("tr",[r("td",[r("code",[e._v("resources.memoryRequestsMissing")])]),e._v(" "),r("td",[r("code",[e._v("warning")])]),e._v(" "),r("td",[e._v("Fails when "),r("code",[e._v("resources.requests.memory")]),e._v(" attribute is not configured.")])]),e._v(" "),r("tr",[r("td",[r("code",[e._v("resources.cpuLimitsMissing")])]),e._v(" "),r("td",[r("code",[e._v("warning")])]),e._v(" "),r("td",[e._v("Fails when "),r("code",[e._v("resources.limits.cpu")]),e._v(" attribute is not configured.")])]),e._v(" "),r("tr",[r("td",[r("code",[e._v("resources.memoryLimitsMissing")])]),e._v(" "),r("td",[r("code",[e._v("warning")])]),e._v(" "),r("td",[e._v("Fails when "),r("code",[e._v("resources.limits.memory")]),e._v(" attribute is not configured.")])])])]),e._v(" "),r("h2",{attrs:{id:"background"}},[r("a",{staticClass:"header-anchor",attrs:{href:"#background"}},[e._v("#")]),e._v(" Background")]),e._v(" "),r("p",[e._v("Configuring resource requests and limits for containers running in Kubernetes is an important best practice to follow. Setting appropriate resource requests will ensure that all your applications have sufficient compute resources. Setting appropriate resource limits will ensure that your applications do not consume too many resources.")]),e._v(" "),r("p",[e._v("Having these values appropriately configured ensures that:")]),e._v(" "),r("ul",[r("li",[r("p",[e._v("Cluster autoscaling can function as intended. New nodes are scheduled once pods are unable to be scheduled on an existing node due to insufficient resources. This will not happen if resource requests are not configured.")])]),e._v(" "),r("li",[r("p",[e._v("Each container has sufficient access to compute resources. Without resource requests, a pod may be scheduled on a node that is already overutilized. Without resource limits, a single poorly behaving pod could utilize the majority of resources on a node, significantly impacting the performance of other pods on the same node.")])])]),e._v(" "),r("h2",{attrs:{id:"further-reading"}},[r("a",{staticClass:"header-anchor",attrs:{href:"#further-reading"}},[e._v("#")]),e._v(" Further Reading")]),e._v(" "),r("ul",[r("li",[r("a",{attrs:{href:"https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/",target:"_blank",rel:"noopener noreferrer"}},[e._v("Kubernetes Docs: Managing Compute Resources for Containers"),r("OutboundLink")],1)]),e._v(" "),r("li",[r("a",{attrs:{href:"https://cloud.google.com/blog/products/gcp/kubernetes-best-practices-resource-requests-and-limits",target:"_blank",rel:"noopener noreferrer"}},[e._v("Kubernetes best practices: Resource requests and limits"),r("OutboundLink")],1)]),e._v(" "),r("li",[r("a",{attrs:{href:"https://github.com/kubernetes/autoscaler/tree/master/vertical-pod-autoscaler",target:"_blank",rel:"noopener noreferrer"}},[e._v("Vertical Pod Autoscaler (can automatically set resource requests and limits)"),r("OutboundLink")],1)])])])}),[],!1,null,null,null);t.default=o.exports}}]);