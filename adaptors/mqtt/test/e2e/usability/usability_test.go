package usability_test

import (
	"encoding/base64"
	"fmt"
	"path/filepath"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/rancher/octopus/adaptors/mqtt/api/v1alpha1"
	edgev1alpha1 "github.com/rancher/octopus/api/v1alpha1"
	"github.com/rancher/octopus/test/util/content"
	"github.com/rancher/octopus/test/util/exec"
	"github.com/rancher/octopus/test/util/node"
)

var (
	testDeviceLink            edgev1alpha1.DeviceLink
	testDeviceLinkName		  = "test-mqtt-"
	publishedMessage          = "publish-message"
	subscribedMessage         string
	temporaryMessage          string
	MQTTServerURL             = "tcp://mosquitto.default:1883"
	templatedTopicToSubscribe = "cattle.io/octopus/home/set/kitchen/door/switch"
)

var _ = Describe("verify usability", func() {

	BeforeEach(func() {
		deployMQTTDeviceLink()
	})

	AfterEach(func() {
		_ = k8sCli.DeleteAllOf(testCtx, &edgev1alpha1.DeviceLink{}, client.InNamespace(testDeviceLink.Namespace))
	})

	Context("modify MQTT device link spec", func() {

		Specify("if invalid node spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid node spec", invalidNodeSpec)

			By("then node of the device link is not found", isNodeExistedFalse)

			By("when correct node spec", correctNodeSpec)

			By("then node of the device link is found", isNodeExistedTrue)

		})

		Specify("if invalid model spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid model spec", invalidModelSpec)

			By("then model of the device link is not found", isModelExistedFalse)

			By("when correct model spec", correctModelSpec)

			By("then model of the device link is found", isModelExistedTrue)

		})

		Specify("if invalid adaptor spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid adaptor spec", invalidAdaptorSpec)

			By("then adaptor of the device link is not found", isAdaptorExistedFalse)

			By("when correct adaptor spec", correctAdaptorSpec)

			By("then adaptor of the device link is found", isAdaptorExistedTrue)

		})

		Specify("if invalid device spec", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid device spec", invalidDeviceSpec)

			By("then the device link is not connected", isDeviceConnectedFalse)

			By("when correct device spec", correctDeviceSpec)

			By("then the device link is connected", isDeviceConnectedTrue)

		})
	})

	Context("publish/subscribe messages through MQTT device link", func() {

		Specify("publish messages through an MQTT device link", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when publish a message to corresponding MQTT server", publishMessage)

			By("then the message should be inside device link", isMessageInsideDeviceLink)

		})

		Specify("subscribe messages through an MQTT device link", func() {

			By("when deploy an MQTT device link with remote server", deployDeviceLinkWithRemoteServer)

			By("then the device link is connected", isDeviceConnectedTrue)

			By("when subscribe a message from server", subscribeMessage)

			By("then the message is same to the value inside device link", isSubscribedMessageValid)

		})

		Specify("subscribe messages through an MQTT device link which does not exist", func() {

			By("when deploy MQTT device link with remote server", deployDeviceLinkWithRemoteServer)

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when subscribe a message and save it", subscribeMessageAndSave)

			By("when delete MQTT device link", deleteMQTTDeviceLink)

			By("when subscribe a message again", subscribeMessage)

			By("then the two messages are the same", isMessagesTheSame)

		})

		Specify("subscribe messages through an MQTT device link whose payload is null", func() {

			By("when deploy an MQTT device link with payload null", deployDeviceLinkWithPayloadNull)

			By("then the device link is connected", isDeviceConnectedTrue)

			By("when subscribe a message", subscribeMessage)

			By("then the message is null", isSubscribeMessageNull)

		})

		Specify("if the MQTT server is unavailable", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when invalid server URL", invalidServerURL)

			By("then the device link is not connected", isDeviceConnectedFalse)

			By("when correct server URL", correctServerURL)

			By("then the device link is connected", isDeviceConnectedTrue)

		})

		Specify("subscribe messages through an MQTT device link whose payload is a complex JSON", func() {

			By("when deploy an MQTT device link whose payload is a complex JSON", deployDeviceLinkWithComplexJSON)

			By("then the device link is connected", isDeviceConnectedTrue)

			By("when subscribe a message", subscribeMessage)

			By("then the subscribed message is same to the JSON value", isSubscribedMessageValid)

		})
	})

	Specify("if subscribe messages of templated topic", func() {

		By("when deploy dummy device link with templated topic", deployDummyDeviceWithTemplatedTopic)

		By("then the device link is connected", isDeviceConnectedTrue)

		By("then succeed subscribing a message from topic", succeedSubscribingTemplatedTopic)
	})

	Context("operation on pods/nodes", func() {

		Specify("if delete adaptor pods", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete MQTT adaptor pods", deleteMQTTAdaptorPods)

			By("then of the device link AdaptorExisted false", isAdaptorExistedFalse)

		})

		Specify("if delete limbs pods", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete limbs pods", deleteLimbsPods)

			By("then the MQTT adaptor pods become error", isMQTTAdaptorPodsError)

		})

		Specify("if delete MQTT device model", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete MQTT device model", deleteMQTTDeviceCRD)

			By("then model of the device link is not found", isModelExistedFalse)

			By("redeploy MQTT device model", redeployMQTTDeviceCRD)

		})

		Specify("if delete corresponding node", func() {

			By("given the device link is connected", isDeviceConnectedTrue)

			By("when delete corresponding cluster node", deleteCorrespondingNode)

			By("then node of the device link is not found", isNodeExistedFalse)
		})
	})
})

type judgeFunc func(edgev1alpha1.DeviceLink) bool

func doDeviceLinkJudgment(judge judgeFunc) {
	Eventually(func() bool {
		var deviceLinkKey = types.NamespacedName{
			Name:      testDeviceLink.Name,
			Namespace: testDeviceLink.Namespace,
		}
		if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
			Fail(err.Error())
		}
		return judge(testDeviceLink)
	}, 300, 3).Should(BeTrue())
}

func getDeviceLinkPointer() *edgev1alpha1.DeviceLink {
	var deviceLinkKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, deviceLinkKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	return &testDeviceLink
}

func deployMQTTDeviceLink() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	testDeviceLink = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    "default",
			GenerateName: testDeviceLinkName,
		},
		Spec: edgev1alpha1.DeviceLinkSpec{
			Adaptor: edgev1alpha1.DeviceAdaptor{
				Node: targetNode,
				Name: "adaptors.edge.cattle.io/mqtt",
			},
			Model: metav1.TypeMeta{
				Kind:       "MQTTDevice",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			},
			Template: edgev1alpha1.DeviceTemplateSpec{
				Spec: content.ToRawExtension(
					map[string]interface{}{
						"properties": []map[string]interface{}{
							{
								"name":     "switch",
								"type":     "boolean",
								"readOnly": false,
								"value":    "false",
							},
						},
						"protocol": map[string]interface{}{
							"pattern": "AttributedTopic",
							"client": map[string]interface{}{
								"server": MQTTServerURL,
							},
							"message": map[string]interface{}{
								"topic": "cattle.io/octopus/home/:operator/kitchen/door/:path",
								"operator": map[string]interface{}{
									"read":  "status",
									"write": "set",
								},
							},
						},
					},
				),
			},
		},
	}
	Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
}

func deployDeviceLinkWithRemoteServer() {
	MQTTServerURL = "tcp://test.mosquitto.org:1883"
	deployMQTTDeviceLink()
	MQTTServerURL = "tcp://mosquitto.default:1883"
}

func deployDeviceLinkWithPayloadNull() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	testDeviceLink = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    testDeviceLink.Namespace,
			GenerateName: testDeviceLinkName,
		},
		Spec: edgev1alpha1.DeviceLinkSpec{
			Adaptor: edgev1alpha1.DeviceAdaptor{
				Node: targetNode,
				Name: "adaptors.edge.cattle.io/mqtt",
			},
			Model: metav1.TypeMeta{
				Kind:       "MQTTDevice",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			},
			Template: edgev1alpha1.DeviceTemplateSpec{
				Spec: content.ToRawExtension(
					map[string]interface{}{
						"properties": []map[string]interface{}{
							{
								"name":     "switch",
								"type":     "boolean",
								"readOnly": false,
								// "value": "false",
							},
						},
						"protocol": map[string]interface{}{
							"pattern": "AttributedTopic",
							"client": map[string]interface{}{
								"server": "tcp://test.mosquitto.org:1883",
							},
							"message": map[string]interface{}{
								"topic": "cattle.io/octopus/home/:operator/kitchen/door/:path",
								"operator": map[string]interface{}{
									"read":  "status",
									"write": "set",
								},
							},
						},
					},
				),
			},
		},
	}
	Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
}

func deployDeviceLinkWithComplexJSON() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	testDeviceLink = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    testDeviceLink.Namespace,
			GenerateName: testDeviceLinkName,
		},
		Spec: edgev1alpha1.DeviceLinkSpec{
			Adaptor: edgev1alpha1.DeviceAdaptor{
				Node: targetNode,
				Name: "adaptors.edge.cattle.io/mqtt",
			},
			Model: metav1.TypeMeta{
				Kind:       "MQTTDevice",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			},
			Template: edgev1alpha1.DeviceTemplateSpec{
				Spec: content.ToRawExtension(
					map[string]interface{}{
						"properties": []map[string]interface{}{
							{
								"name":     "switch",
								"type":     "boolean",
								"readOnly": false,
								"value": map[string]interface{}{
									"name":      map[string]interface{}{"first": "Tom", "last": "Anderson"},
									"age":       37,
									"children":  []string{"Sara", "Alex", "Jack"},
									"fav.movie": "Deer Hunter",
									"friends": []interface{}{
										map[string]interface{}{"first": "Dale", "last": "Murphy", "age": 44, "nets": []string{"ig", "fb", "tw"}},
										map[string]interface{}{"first": "Roger", "last": "Craig", "age": 68, "nets": []string{"fb", "tw"}},
										map[string]interface{}{"first": "Jane", "last": "Murphy", "age": 47, "nets": []string{"ig", "tw"}},
									},
								},
							},
						},
						"protocol": map[string]interface{}{
							"pattern": "AttributedTopic",
							"client": map[string]interface{}{
								"server": "tcp://test.mosquitto.org:1883",
							},
							"message": map[string]interface{}{
								"topic": "cattle.io/octopus/home/:operator/kitchen/door/:path",
								"operator": map[string]interface{}{
									"read":  "status",
									"write": "set",
								},
							},
						},
					},
				),
			},
		},
	}
	Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
}

func deployDummyDeviceWithTemplatedTopic() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	testDeviceLink = edgev1alpha1.DeviceLink{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:    testDeviceLink.Namespace,
			GenerateName: testDeviceLinkName,
		},
		Spec: edgev1alpha1.DeviceLinkSpec{
			Adaptor: edgev1alpha1.DeviceAdaptor{
				Node: targetNode,
				Name: "adaptors.edge.cattle.io/dummy",
			},
			Model: metav1.TypeMeta{
				Kind:       "DummySpecialDevice",
				APIVersion: "devices.edge.cattle.io/v1alpha1",
			},
			Template: edgev1alpha1.DeviceTemplateSpec{
				Spec: content.ToRawExtension(
					map[string]interface{}{
						"extension": map[string]interface{}{
							"mqtt": map[string]interface{}{
								"client": map[string]interface{}{
									"server": "tcp://test.mosquitto.org:1883",
								},
								"message": map[string]interface{}{
									"topic": "octopus/:namespace/:name/:operator/:path",
									"operator": map[string]interface{}{
										"read":  "status",
										"write": "set",
									},
									"path": "yosemite",
								},
							},
						},
						"gear": "slow",
						"on":   true,
						"protocol": map[string]interface{}{
							"location": "home",
						},
					},
				),
			},
		},
	}
	Expect(k8sCli.Create(testCtx, &testDeviceLink)).Should(Succeed())
}

func correctNodeSpec() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(fmt.Sprintf(`{"spec":{"adaptor":{"node":"%s"}}}`, targetNode))
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidNodeSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"adaptor":{"node":"wrong-node"}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isNodeExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetNodeExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isNodeExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetNodeExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctModelSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"model":{"apiVersion":"devices.edge.cattle.io/v1alpha1"}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidModelSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"model":{"apiVersion":"wrong-apiVersion"}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isModelExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetModelExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isModelExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetModelExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctAdaptorSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"adaptor":{"name":"adaptors.edge.cattle.io/mqtt"}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidAdaptorSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"adaptor":{"name":"wrong-adaptor-name"}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isAdaptorExistedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetAdaptorExistedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isAdaptorExistedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetAdaptorExistedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func correctDeviceSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"template":{"spec":{"protocol":{"pattern":"AttributedTopic"}}}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func invalidDeviceSpec() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"template":{"spec":{"protocol":{"pattern":"AttributedTopic---"}}}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func isDeviceConnectedTrue() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceConnectedStatus() == metav1.ConditionTrue
	}
	doDeviceLinkJudgment(judge)
}

func isDeviceConnectedFalse() {
	var judge = func(deviceLink edgev1alpha1.DeviceLink) bool {
		return deviceLink.GetDeviceConnectedStatus() == metav1.ConditionFalse
	}
	doDeviceLinkJudgment(judge)
}

func deleteMQTTAdaptorPods() {
	Expect(k8sCli.DeleteAllOf(testCtx, &corev1.Pod{}, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-mqtt"})).
		Should(Succeed())
}

func deleteLimbsPods() {
	Expect(k8sCli.DeleteAllOf(testCtx, &corev1.Pod{}, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/component": "limb"})).
		Should(Succeed())
}

func isMQTTAdaptorPodsError() {
	var podList corev1.PodList
	Eventually(func() bool {
		if err := k8sCli.List(testCtx, &podList, client.InNamespace("octopus-system"), client.MatchingLabels{"app.kubernetes.io/name": "octopus-adaptor-mqtt"}); err != nil {
			Fail(err.Error())
		}
		for _, pod := range podList.Items {
			for _, condition := range pod.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "False" {
					return true
				}
			}
		}
		return false
	}, 300, 1).Should(BeTrue())
}

func deleteCorrespondingNode() {
	var targetNode, err = node.GetValidWorker(testCtx, k8sCli)
	Expect(err).ShouldNot(HaveOccurred())

	var correspondingNode = corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: targetNode,
		},
	}
	Expect(k8sCli.Delete(testCtx, &correspondingNode)).Should(Succeed())
}

func deleteMQTTDeviceCRD() {
	var crd = v1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "mqttdevices.devices.edge.cattle.io",
		},
	}
	Expect(k8sCli.Delete(testCtx, &crd)).Should(Succeed())
}

func redeployMQTTDeviceCRD() {
	Expect(exec.RunKubectl(nil, GinkgoWriter, "apply", "-f", filepath.Join(testCurrDir, "deploy", "manifests", "crd", "base", "devices.edge.cattle.io_mqttdevices.yaml"))).
		Should(Succeed())
}

func invalidServerURL() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"template":{"spec":{"protocol":{"client":{"server":"wrong-server"}}}}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func correctServerURL() {
	deviceLinkPtr := getDeviceLinkPointer()
	patch := []byte(`{"spec":{"template":{"spec":{"protocol":{"client":{"server":"tcp://mosquitto.default:1883"}}}}}}`)
	Expect(k8sCli.Patch(testCtx, deviceLinkPtr, client.RawPatch(types.MergePatchType, patch))).Should(Succeed())
}

func publishMessage() {
	var podList corev1.PodList
	Expect(k8sCli.List(testCtx, &podList, client.InNamespace("default"))).Should(Succeed())
	for _, pod := range podList.Items {
		Expect(exec.RunKubectl(nil, GinkgoWriter, "exec", pod.Name, "--", "mosquitto_pub", "-h", "mosquitto",
			"-t", "cattle.io/octopus/home/status/kitchen/door/switch", "-m", publishedMessage)).Should(Succeed())
	}
}

func isMessageInsideDeviceLink() {
	var device v1alpha1.MQTTDevice
	count := 0
	Eventually(func() bool {
		count++
		var targetKey = types.NamespacedName{
			Namespace: testDeviceLink.Namespace,
			Name:      testDeviceLink.Name,
		}
		if err := k8sCli.Get(testCtx, targetKey, &device); err != nil {
			Fail(err.Error())
		}
		if len(device.Status.Properties) > 0 {
			return string(device.Status.Properties[0].Value.Raw) == "\""+base64.StdEncoding.EncodeToString([]byte(publishedMessage))+"\""
		}
		if count%10 == 0 && len(device.Status.Properties) == 0 {
			publishMessage()
		}
		return false
	}, 300, 1).Should(BeTrue())
}

func subscribeMessage() {
	opts := MQTT.NewClientOptions()
	opts.AddBroker("tcp://test.mosquitto.org:1883")

	choke := make(chan [2]string)

	opts.SetDefaultPublishHandler(func(client MQTT.Client, msg MQTT.Message) {
		choke <- [2]string{msg.Topic(), string(msg.Payload())}
	})

	client0 := MQTT.NewClient(opts)
	if token := client0.Connect(); token.Wait() && token.Error() != nil {
		Fail(token.Error().Error())
	}

	topic := templatedTopicToSubscribe
	if token := client0.Subscribe(topic, byte(1), nil); token.Wait() && token.Error() != nil {
		Fail(token.Error().Error())
	}

	incoming := <-choke
	subscribedMessage = incoming[1]

	client0.Disconnect(250)
}

func isSubscribedMessageValid() {
	var targetKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	var device v1alpha1.MQTTDevice
	Eventually(func() bool {
		if err := k8sCli.Get(testCtx, targetKey, &device); err != nil {
			Fail(err.Error())
		}
		return string(device.Spec.Properties[0].Value.Raw) == subscribedMessage
	}, 300, 1).Should(BeTrue())
}

func isMessagesTheSame() {
	Expect(temporaryMessage == subscribedMessage).Should(BeTrue())
}

func subscribeMessageAndSave() {
	subscribeMessage()
	temporaryMessage = subscribedMessage
	subscribedMessage = ""
}

func deleteMQTTDeviceLink() {
	var targetKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	if err := k8sCli.Get(testCtx, targetKey, &testDeviceLink); err != nil {
		Fail(err.Error())
	}
	Expect(k8sCli.Delete(testCtx, &testDeviceLink)).Should(Succeed())
}

func isSubscribeMessageNull() {
	var targetKey = types.NamespacedName{
		Namespace: testDeviceLink.Namespace,
		Name:      testDeviceLink.Name,
	}
	var device v1alpha1.MQTTDevice
	Eventually(func() bool {
		if err := k8sCli.Get(testCtx, targetKey, &device); err != nil {
			Fail(err.Error())
		}
		return "null" == subscribedMessage
	}, 300, 1).Should(BeTrue())
}

func succeedSubscribingTemplatedTopic() {
	templatedTopicToSubscribe = fmt.Sprintf("octopus/%s/%s/set/yosemite", testDeviceLink.Namespace, testDeviceLink.Name)
	subscribedMessage = ""
	// actually the subscribed message varies
	subscribeMessage()
	templatedTopicToSubscribe = "cattle.io/octopus/home/set/kitchen/door/switch"

	Expect(subscribedMessage != "").Should(BeTrue())
}
