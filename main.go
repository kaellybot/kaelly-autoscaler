package main

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

func main() {
	token := os.Getenv("TOKEN")
	namespace := os.Getenv("NAMESPACE")
	statefulSetName := os.Getenv("STATEFULSET")
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		panic(err)
	}
	st, err := dg.GatewayBot()
	if err != nil {
		log.Fatal().Err(err).
			Msgf("Impossible to connect to Discord, shutting down...")
	}

	log.Info().Msgf("Recommended Shards: %v", st.Shards)
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal().Err(err).
			Msgf("Impossible to retrieve cluster config, shutting down...")
	}
	// creates the clientset
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	statefulSetsClient := clientSet.AppsV1().StatefulSets(namespace)

	result, err := statefulSetsClient.Get(context.Background(), statefulSetName, metav1.GetOptions{})
	if err != nil {
		log.Fatal().Err(err).
			Msgf("Failed to getStatefulSet, shutting down...")
	}

	if st.Shards > math.MaxInt32 || st.Shards < math.MinInt32 {
		log.Fatal().Err(err).
			Msgf("Shard Count value cannot be casted to int32, shutting down...")
	}
	//nolint:gosec // False positive.
	shardCount := int32(st.Shards)
	if shardCount > *result.Spec.Replicas {
		result = updateReplicas(result, shardCount)
		_, err = statefulSetsClient.Update(context.Background(), result, metav1.UpdateOptions{})
		if err != nil {
			log.Fatal().Err(err).
				Msgf("Failed to update StatefulSet, shutting down...")
		}

		log.Info().
			Msgf("Updated StatefulSet to %v replicas", st.Shards)

		// TODO rabbitmq notification?
	}
}

func updateReplicas(result *v1.StatefulSet, shardCount int32) *v1.StatefulSet {
	containerName := os.Getenv("CONTAINER_NAME")
	shardIDEnvVar := os.Getenv("SHARD_ID_ENV_VAR")
	shardCountEnvVar := os.Getenv("SHARD_COUNT_ENV_VAR")

	// Set the replica count
	result.Spec.Replicas = &shardCount
	// Set the shard ID and count environment variables - will force existing shards to restart
	for i, container := range result.Spec.Template.Spec.Containers {
		if container.Name == containerName {
			for j, environmentVariable := range result.Spec.Template.Spec.Containers[i].Env {
				if environmentVariable.Name == shardIDEnvVar {
					result.Spec.Template.Spec.Containers[i].Env[j].Value = strconv.Itoa(i)
				}
				if environmentVariable.Name == shardCountEnvVar {
					result.Spec.Template.Spec.Containers[i].Env[j].Value = fmt.Sprintf("%v", shardCount)
				}
			}
		}
	}

	return result
}
