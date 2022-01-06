# StrategyExecutor

#### Intro
StrategyExecutor is a gitee event massage executor, it recieves rules and events pushes from datacache, and do the commands.

#### Architect
![](http://assets.processon.com/chart_image/6163e23e0791290cc7819291.png)

#### Domain Model
![](http://assets.processon.com/chart_image/616428d163768921fa176b05.png)

#### How to contribute
If you’re interested in contributing code, the best starting point is to have a look at our Gitee issues to see which tasks are the most urgent. 

Sunmao accepts PR's (pull requests) from all developers.

Issues can be submitted by anyone - either seasoned developers or newbies.

#### Installation

- **Step 1** Setting up the k8s environment, Google GKE or minikube or microk8s are ok for deployment.

- **Step 2** Setting up webhook url in gitee projects, in order to receive issue event requests from you project.

- **Step 3** Setting up `api_url`, `Org`, `Repo`, and `gitee_token` environment variables.

- **Step 4** If you use RabbitMQ as a message queue, you should set these environment variables:
```
	RMQ_QUEUE_NAME := os.Getenv("RMQ_QUEUE_NAME")
	RMQ_HOST := os.Getenv("RMQ_HOST")
	RMQ_VHOST := os.Getenv("RMQ_VHOST")
	RMQ_USER := os.Getenv("RMQ_USER")
	RMQ_PASS := os.Getenv("RMQ_PASS")
	RMQ_PORT := os.Getenv("RMQ_PORT")
	RMQ_ROUTINGKEY := os.Getenv("RMQ_ROUTINGKEY")
	RMQ_EXCHANGE_NAME := os.Getenv("RMQ_EXCHANGE_NAME")
	RMQ_EXCHANGE_TYPE := os.Getenv("RMQ_EXCHANGE_TYPE")
```
and then setting up a RabbitMQ environment by yourself.


- **Step 5** Using Dockerfile to build docker image and then upload it to DockerHub.

- **Step 6** Deploy the project by yaml on k8s.