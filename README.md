# kaelly-autoscaler

Discord bot shard autoscaler for Kubernetes, written in Go

In details, it is a K3s/K8s StatefulSet autoscaler that uses the Discord Gateway Bot endpoint to determine the recommended number of shards and automatically scale up when Discord recommends it.

This project requires that you already be using Kubernetes, and assume you have some understand of how Kubernetes works. It also assumes that you have your bot set up to handle changes in the StatefulSet's replica count gracefully. Meaning: if we scale up, all existing shards will need to re-identify with Discord to present the new shard count, and update their local cache as necessary.