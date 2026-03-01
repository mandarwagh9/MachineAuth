package services

import (
	"log"
	"sync"
	"time"
)

// DeliveryWorker processes pending webhook deliveries with a worker pool
type DeliveryWorker struct {
	webhookService *WebhookService
	workerCount    int
	pollInterval   time.Duration
	stopCh         chan struct{}
	wg             sync.WaitGroup
}

// NewDeliveryWorker creates a new delivery worker
func NewDeliveryWorker(webhookService *WebhookService, workerCount int) *DeliveryWorker {
	if workerCount <= 0 {
		workerCount = 3
	}
	return &DeliveryWorker{
		webhookService: webhookService,
		workerCount:    workerCount,
		pollInterval:   10 * time.Second,
		stopCh:         make(chan struct{}),
	}
}

// Start begins the delivery worker pool
func (w *DeliveryWorker) Start() {
	log.Printf("starting webhook delivery worker pool with %d workers", w.workerCount)

	for i := 0; i < w.workerCount; i++ {
		w.wg.Add(1)
		go w.worker(i)
	}
}

// Stop gracefully shuts down the delivery worker
func (w *DeliveryWorker) Stop() {
	log.Println("stopping webhook delivery worker pool...")
	close(w.stopCh)
	w.wg.Wait()
	log.Println("webhook delivery worker pool stopped")
}

func (w *DeliveryWorker) worker(id int) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processPendingDeliveries(id)
		}
	}
}

func (w *DeliveryWorker) processPendingDeliveries(workerID int) {
	deliveries, err := w.webhookService.db.ListPendingDeliveries()
	if err != nil {
		log.Printf("worker %d: failed to list pending deliveries: %v", workerID, err)
		return
	}

	for _, delivery := range deliveries {
		select {
		case <-w.stopCh:
			return
		default:
		}

		webhook, err := w.webhookService.db.GetWebhook(delivery.WebhookConfigID)
		if err != nil {
			log.Printf("worker %d: webhook not found for delivery %s: %v", workerID, delivery.ID, err)
			continue
		}

		if !webhook.IsActive {
			continue
		}

		// Process pending deliveries that are ready for retry
		go w.webhookService.processDelivery(delivery.ID, *webhook, delivery)
	}
}
