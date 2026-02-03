## Future Improvements

Core Backend and Distributed Systems (SDE)
- Optimized a monolithic backend into a high-concurrency microservices architecture via gRPC, resulting in a 40% reduction in inter-service latency
- Engineered a real-time dispatch engine using Apache Pulsar for load-balanced task queues and Redis for proximity indexing, enabling sub-50ms vehicle allocation for thousands of concurrent orders
- Designed a scalable cloud infrastructure on Google Kubernetes Engine (GKE), utilizing Terraform for Infrastructure as Code (IaC) to provision ephemeral environments and Helm to manage stateful workloads like Pulsar and PostgreSQL
Data Engineering and AI Integration (AI/ML)
- Built a real-time ELT pipeline using Apache Pulsar and TimescaleDB, ingesting high-frequency drone telemetry (100+ events/sec) to facilitate long-term historical analysis and model training
- Developed an AI-driven routing service by integrating a Python inference microservice with the Go backend, utilizing Graph Neural Networks (GNN) to optimize delivery paths based on historical traffic topology
- Implemented a Retrieval-Augmented Generation (RAG) system for unstructured command parsing, deploying a local LLM alongside the microservices cluster to securely process natural language dispatch instructions

### AI/ML Improvements

Option A: The "Real" Engineering Path (Route Optimization)
- The Tech: Reinforcement Learning (RL) or Graph Neural Networks (GNN).
- The Logic: Don't use an LLM for routing. Use a GNN to predict "Estimated Time of Arrival" (ETA) based on traffic topology, or use Deep Q-Learning (DQN) to assign orders to drones.
- Why it's better:
  - Latency: A small RL model runs in milliseconds.
  - Resume Value: Shows you know PyTorch/TensorFlow, model serving, and math.
- Implementation:
  - Train a simple agent using Stable Baselines3 (Python) on a simulated grid.
  - Export the model to ONNX.
  - Load it in your Python microservice.

Option B: The "LLM" Path (But Engineered Correctly)
- The Tech: Local LLM (SLM - Small Language Model) like Llama-3-8B or Phi-3 (Microsoft).
- The Logic: Use the LLM for Unstructured Data Parsing. For example, a user types "Leave it behind the red potting plant near the gate." The LLM extracts { "drop_location": "gate", "landmark": "red plant" }.
- Why it's better:
  - Privacy/Cost: No data leaves your server.
  - Engineering: You have to manage GPU memory, quantization (running 4-bit models), and containerization (Docker with CUDA).
- Implementation:
  - Use Ollama or vLLM to host the model locally alongside your Go services.
  - Your Go service calls this local endpoint.
        
.
├── cmd
│   └── ... (Your Go services)
├── internal
│   └── ...
├── services          <-- NEW: Distinct folder for non-Go services
│   └── ai-engine     <-- Python Microservice
│       ├── model/    <-- Saved .pt or .onnx files
│       ├── src/
│       │   ├── inference.py
│       │   └── main.py  <-- gRPC Server
│       ├── Dockerfile
│       └── requirements.txt
├── proto
│   └── ai_engine
│       └── inference.proto <-- Defines input (State) and output (Action)

## Workflow Diagram

[ Drones / Robots ]
              │
              ▼ (gRPC / MQTT)
      [ Gateway Service ]
              │
              ├─── (Hot Data: "Where am I now?") ──▶ [ REDIS ] (Geo Index)
              │                                          ▲
              │                                          │ (Query for Dispatch)
              ▼ (Async: "Log my path")                   │
      [ PULSAR CLUSTER ] ────────────────────────┐       │
              │                                  │       │
              ▼                                  ▼       ▼
    [ Tracking Service ]                 [ Dispatch Service ]
    (Writes to TimescaleDB)              (The AI / RL Model)
                                                 │
                                                 ▼
                                        [ Order Service ]
                                        (Writes to Postgres)

## Microservices Brief Summaries

- users: auth and profiles
- orders: state machine for orders (pending -> assigned -> delivered)
- items: inventory management
- fleet: drone / robot fleet management
- dispatch: decision maker
- tracking: high-speed data ingestor
- notification: decouple 3rd party API failures
- recommendation (ai-engine): inference engine for route options
