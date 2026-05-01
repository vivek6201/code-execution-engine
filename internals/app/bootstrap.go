package app

import (
	"context"
	"log"

	"github.com/code-execution-engine/config"
	"github.com/code-execution-engine/internals/api/routes"
	"github.com/code-execution-engine/internals/core/evaluator"
	"github.com/code-execution-engine/internals/core/execution"
	"github.com/code-execution-engine/internals/engine/executor"
	"github.com/code-execution-engine/internals/engine/runners"
	"github.com/code-execution-engine/internals/engine/runners/languages"
	"github.com/code-execution-engine/internals/infra/cache"
	"github.com/code-execution-engine/internals/infra/isolation"
	"github.com/code-execution-engine/internals/infra/queue"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func StartServer() {
	cfg := config.Load()
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	q := queue.NewRedisQueue(redisClient)

	r := gin.Default()
	r.Use(cors.Default())
	routes.SetupRoutes(r, q, redisClient)

	r.Run(":8080")
}

func StartWorker() {
	cfg := config.Load()
	redisClient := cache.NewRedisClient(cfg.RedisUrl)
	q := queue.NewRedisQueue(redisClient)

	containerClient := isolation.NewClient()

	factory := runners.NewFactory()
	factory.Register("python", languages.NewPythonRunner(containerClient))
	factory.Register("cpp", languages.NewCppRunner(containerClient))
	factory.Register("java", languages.NewJavaRunner(containerClient))

	exec := executor.NewExecutor(factory)
	eval := evaluator.NewService()
	service := execution.NewService(exec, eval)

	log.Println("Worker started. Listening for jobs...")

	ctx := context.Background()
	for {
		jobID, j, err := q.Dequeue(ctx)
		if err != nil {
			log.Printf("Error dequeueing job: %v", err)
			continue
		}

		log.Printf("Executing job %s", jobID)
		res := service.Execute(j)

		err = queue.SetResult(ctx, redisClient, jobID, res)
		if err != nil {
			log.Printf("Error saving result for job %s: %v", jobID, err)
		} else {
			log.Printf("Job %s finished", jobID)
		}
	}
}
