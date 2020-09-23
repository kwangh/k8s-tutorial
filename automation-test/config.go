package main

import "time"

const (
	// podStartTimeout is how long to wait for the pod to be started.
	// Initial pod start can be delayed O(minutes) by slow docker pulls.
	// TODO: Make this 30 seconds once #4566 is resolved.
	podStartTimeout = 30 * time.Second
	// PodDeleteTimeout is how long to wait for a pod to be deleted.
	PodDeleteTimeout = 5 * time.Minute

	// poll is how often to poll pods, nodes and claims.
	poll = 2 * time.Second

	// pollLongTimeout is time limit to check deployment to be completed
	pollShortTimeout = 1 * time.Minute
	// pollLongTimeout is time limit to check deployment to be completed
	pollLongTimeout = 5 * time.Minute
)
