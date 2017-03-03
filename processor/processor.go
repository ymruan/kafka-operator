package processor

import (
	k8sclient "k8s.io/client-go/kubernetes"
	kafkaOperatorSpec "github.com/krallistic/kafka-operator/spec"
	"fmt"
	"github.com/krallistic/kafka-operator/util"
)

type Processor struct {
	client k8sclient.Clientset
	baseBrokerImage string
	util util.ClientUtil
	kafkaClusters map[string]*kafkaOperatorSpec.KafkaCluster
}

func New(client k8sclient.Clientset, image string, util util.ClientUtil) (*Processor, error) {
	p := &Processor{
		client:client,
		baseBrokerImage:image,
		util:util,
		kafkaClusters:make(map[string]*kafkaOperatorSpec.KafkaCluster),
	}
	fmt.Println("Created Processor")
	return p, nil
}

func ( p *Processor) Run() error {
	//TODO getListOfAlredyRunningCluster/Refresh
	fmt.Println("Running Processor")
	return nil
}

func ( p *Processor) WatchKafkaEvents(control chan int) {
	rawEventsChannel, errorChannel := p.util.MonitorKafkaEvents()
	fmt.Println("Watching Kafka Events")
	go func() {
		for {
			select {
			case currentEvent := <- rawEventsChannel:
				fmt.Println("Recieved Raw Event, proceeding: ", currentEvent)
				switch currentEvent.Type {
				case "ADDED":
					fmt.Println("ADDED")
					p.CreateKafkaCluster(currentEvent.Object)
				case "MODIFIED":
					fmt.Println("MODIFIED")
				default:
					fmt.Println(currentEvent.Type)
				}
			case err := <- errorChannel:
				println("Error Channel", err)
			case <-control:
				fmt.Println("Recieved Something on Control Channel, shutting down: ")
				return
			}

		}
	}()


}

func (p *Processor) CreateKafkaCluster(clusterSpec kafkaOperatorSpec.KafkaCluster) {
	fmt.Println("CreatingKafkaCluster", clusterSpec)
	fmt.Println("SPEC: ", clusterSpec.Spec)
	// TODO What happens if connections loss? after a reconnect we get ADDED again :/
	// We need to hold State?


	headless_SVC_Name := clusterSpec.Spec.Name
	suffix := ".cluster.local:9092"
	brokerNames := make([]string, clusterSpec.Spec.Brokers.Count)

	round_robing_dns := headless_SVC_Name + suffix
	fmt.Println("Headless Service Name: ", headless_SVC_Name, " Should be accessable through LB: ", round_robing_dns )

	var i int32
	for  i = 0; i < clusterSpec.Spec.Brokers.Count; i++ {
		brokerNames[i] = "kafka-0." + headless_SVC_Name + suffix
		fmt.Println("Broker", i , " ServiceName: ", brokerNames[i])
	}




	//Create Headless Brokersvc
	//TODO better naming
	p.util.CreateBrokerService(headless_SVC_Name, false)

	//CREATE Broker sts
	//Currently we extract name out of spec, maybe move to metadata to be more inline with other k8s komponents.
	p.util.CreateBrokerStatefulSet(clusterSpec.Spec)





}
