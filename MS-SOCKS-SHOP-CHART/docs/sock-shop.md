# Sock-Shop Application

Sock-Shop is a microservices demo application that simulates an e-commerce website selling socks. It's designed to demonstrate cloud-native technologies and microservices architecture.

## Architecture

```
                                    ┌─────────────┐
                                    │  front-end  │
                                    │   (Node.js) │
                                    └──────┬──────┘
                                           │
           ┌───────────────┬───────────────┼───────────────┬───────────────┐
           │               │               │               │               │
           ▼               ▼               ▼               ▼               ▼
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
    │  catalogue  │ │    carts    │ │   orders    │ │   payment   │ │    user     │
    │    (Go)     │ │   (Java)    │ │   (Java)    │ │    (Go)     │ │    (Go)     │
    └──────┬──────┘ └──────┬──────┘ └──────┬──────┘ └─────────────┘ └──────┬──────┘
           │               │               │                               │
           ▼               ▼               ▼                               ▼
    ┌─────────────┐ ┌─────────────┐ ┌─────────────┐                 ┌─────────────┐
    │catalogue-db │ │  carts-db   │ │  orders-db  │                 │   user-db   │
    │   (MySQL)   │ │  (MongoDB)  │ │  (MongoDB)  │                 │  (MongoDB)  │
    └─────────────┘ └─────────────┘ └─────────────┘                 └─────────────┘
                                           │
                                           ▼
                                    ┌─────────────┐
                                    │  shipping   │───────▶ ┌─────────────┐
                                    │   (Java)    │         │queue-master │
                                    └─────────────┘         │   (Java)    │
                                                            └──────┬──────┘
                                                                   │
                                                                   ▼
                                                            ┌─────────────┐
                                                            │  rabbitmq   │
                                                            └─────────────┘
```

## Services

### Frontend

| Property | Value |
|----------|-------|
| Language | Node.js |
| Port | 8079 |
| Image | `weaveworksdemos/front-end:0.3.12` |
| Service Type | LoadBalancer |

The frontend is the main entry point for users. It provides a web UI for browsing and purchasing socks.

### Catalogue

| Property | Value |
|----------|-------|
| Language | Go |
| Port | 80 |
| Image | `weaveworksdemos/catalogue:0.3.5` |
| Database | MySQL (catalogue-db) |

Manages the product catalog - list of socks, descriptions, prices, and images.

### Carts

| Property | Value |
|----------|-------|
| Language | Java (Spring Boot) |
| Port | 80 |
| Image | `weaveworksdemos/carts:0.4.8` |
| Database | MongoDB (carts-db) |

Handles shopping cart functionality - add/remove items, update quantities.

### Orders

| Property | Value |
|----------|-------|
| Language | Java (Spring Boot) |
| Port | 80 |
| Image | `weaveworksdemos/orders:0.4.7` |
| Database | MongoDB (orders-db) |

Processes customer orders and maintains order history.

### Payment

| Property | Value |
|----------|-------|
| Language | Go |
| Port | 80 |
| Image | `weaveworksdemos/payment:0.4.3` |

Simulates payment processing (always succeeds in demo mode).

### Shipping

| Property | Value |
|----------|-------|
| Language | Java (Spring Boot) |
| Port | 80 |
| Image | `weaveworksdemos/shipping:0.4.8` |

Handles shipping calculations and dispatches orders to the queue.

### User

| Property | Value |
|----------|-------|
| Language | Go |
| Port | 80 |
| Image | `weaveworksdemos/user:0.4.7` |
| Database | MongoDB (user-db) |

Manages user accounts, authentication, and addresses.

### Queue Master

| Property | Value |
|----------|-------|
| Language | Java (Spring Boot) |
| Port | 80 |
| Image | `weaveworksdemos/queue-master:0.3.1` |
| Message Broker | RabbitMQ |

Processes messages from the shipping queue.

### RabbitMQ

| Property | Value |
|----------|-------|
| Image | `rabbitmq:3.6.8` |
| Port | 5672 |

Message broker for async communication between services.

---

## Configuration

### values.yaml

```yaml
sockShop:
  enabled: true
  
  frontEnd:
    replicas: 1
    image: weaveworksdemos/front-end:0.3.12
    resources:
      requests:
        cpu: 100m
        memory: 100Mi
    service:
      type: LoadBalancer
      port: 80
      targetPort: 8079
      nodePort: 30001

  catalogue:
    replicas: 1
    image: weaveworksdemos/catalogue:0.3.5

  carts:
    replicas: 1
    image: weaveworksdemos/carts:0.4.8
    javaOpts: "-Xms64m -Xmx128m -XX:+UseG1GC"

  # ... other services
```

### Scaling Services

```bash
# Scale frontend to 3 replicas
kubectl scale deployment front-end -n sock-shop --replicas=3

# Or update values.yaml and upgrade
helm upgrade sock-shop .
```

---

## Accessing the Application

### Via Minikube (Recommended for Windows)

```bash
minikube service front-end -n sock-shop
```

### Via Port Forward

```bash
kubectl port-forward svc/front-end -n sock-shop 8081:80
# Then open http://localhost:8081
```

### Via NodePort

```bash
# Get minikube IP
minikube ip

# Access via NodePort
# http://<minikube-ip>:30001
```

---

## Default Users

The application comes with pre-seeded test users:

| Username | Password |
|----------|----------|
| user | password |
| user1 | password |
| Eve_Berger | eve |

---

## API Endpoints

### Frontend
- `GET /` - Home page
- `GET /category.html` - Product listing
- `GET /detail.html?id=<id>` - Product detail
- `GET /basket.html` - Shopping cart
- `GET /customer-orders.html` - Order history

### Catalogue API
- `GET /catalogue` - List all products
- `GET /catalogue/{id}` - Get product by ID
- `GET /catalogue/size` - Get total count
- `GET /tags` - Get all tags

### Carts API
- `GET /carts/{customerId}` - Get cart
- `POST /carts/{customerId}` - Add item
- `DELETE /carts/{customerId}` - Clear cart

### Orders API
- `GET /orders` - List orders
- `POST /orders` - Create order

### User API
- `GET /customers` - List users
- `GET /customers/{id}` - Get user
- `POST /register` - Register new user
- `GET /login` - Authenticate

---

## Troubleshooting

### Java Services Restarting

Java services (carts, orders, shipping, queue-master) may restart a few times initially due to memory constraints. This is normal - they stabilize after a few minutes.

```bash
# Check logs
kubectl logs deployment/carts -n sock-shop

# Increase memory if needed (in values.yaml)
carts:
  javaOpts: "-Xms128m -Xmx256m"
```

### Database Connection Issues

```bash
# Check database pods
kubectl get pods -n sock-shop | grep db

# Check database logs
kubectl logs deployment/catalogue-db -n sock-shop
kubectl logs deployment/carts-db -n sock-shop
```

---

## Related Documentation

- [Getting Started](getting-started.md)
- [Prometheus](prometheus.md) - Monitor sock-shop metrics
- [Grafana](grafana.md) - Visualize service health
