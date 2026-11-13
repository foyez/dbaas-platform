# Building a Cloud-Native Platform: From IaaS to PaaS

**Meta-Track: Infrastructure → Platform → Product**

## [Week 0 - Prerequisites for this Course](./week-0/README.md)

## [Week 1 – IaaS Foundations & Provisioning a Kubernetes Cluster on OpenStack](./week-1/README.md)

You start at the bottom, no shortcuts.

### Goal
Understand and operate Infrastructure-as-a-Service (IaaS) by setting up and validating an OpenStack environment.

### Scope
1. **Understanding the OpenStack Architecture**
   - [x] How OpenStack is structured
   - Responsibilities of core components: 
        - [x] Virtual Machines (Nova)
        - [x] Networks (Neutron)
        - [x] Storage (Cinder and Swift), and related services
2. **Installation and Basic Configuration** of a single-node OpenStack cluster using DevStack
   - [x] Install DevStack on a supported Linux distribution
   - [x] Configure the `local.conf` file
   - [x] Deploy a single-node OpenStack environment
   - [x] Verify that all required OpenStack services are running
3. **Validation of the Installation** by:
   - [x] Check OpenStack system logs for errors
   - [x] Verify the status and health of OpenStack services
   - [x] Confirm access to the Horizon dashboard and OpenStack CLI
4. **Deployment** of a virtual machine via OpenStack
   - [x] Upload a VM image (if required)
   - [x] Create the required network resources
   - [x] Launch a virtual machine instance
   - [ ] Verify that the instance is reachable and operational
5. **Architecture Analysis:**
   - [ ] Analyze the deployed OpenStack environment
   - [ ] Identify the roles of the deployed services
   - [ ] Create a technically accurate architecture diagram of the installation
6. **Understanding Terraform and Infrastructure as Code (IaC):**
   - [ ] What Infrastructure as Code is and its core advantages
   - [ ] Terraform basics: Providers, Resources, State
   - [ ] The core Terraform workflow (`init`, `plan`, and `apply`)
7. **Provision a Virtual Machine** with Terraform using the OpenStack APIs
   - [ ] Configure the OpenStack Terraform provider
   - [ ] Authenticate Terraform against the OpenStack APIs
   - [ ] Define the infrastructure required for a virtual machine
   - [ ] Provision the virtual machine using Terraform
   - [ ] Verify that the virtual machine was created successfully in OpenStack
8. **Installation and Configuration of Kubernetes** with a tool of your choice
   - [ ] Select a Kubernetes installation tool (e.g., kubeadm, Minikube, or k3s).
   - [ ] Install a Kubernetes cluster.
   - [ ] Verify that the control plane and worker node(s) are operational.
   - [ ] Deploy a test application to validate the cluster.

### Bonus
- **Fully Automated Installation:** Implement a single command to automatically provision the OpenStack virtual machine and install the Kubernetes cluster.
   - [ ] Automate the provisioning of an OpenStack virtual machine.
   - [ ] Automate the installation and configuration of a Kubernetes cluster.
   - [ ] Provide a single command that performs the complete deployment.
   - [ ] Verify that the virtual machine and Kubernetes cluster are operational after deployment.
- **Two-Node Kubernetes Cluster Setup:** Expand the installation script to provision two virtual machines and configure them to form a two-node Kubernetes cluster.
   - [ ] Extend the automation to provision two OpenStack virtual machines.
   - [ ] Configure one virtual machine as the Kubernetes control plane.
   - [ ] Configure the second virtual machine as a Kubernetes worker node.
   - [ ] Verify that both nodes successfully form a Kubernetes cluster.
   - [ ] Validate cluster functionality by deploying a test workload.

---


## [Week 2 – Developing a Platform-as-a-Service (PaaS) Product on STACKIT Kubernetes Engine (SKE)](./week-2/README.md)

Infrastructure is useless without a product. Now you build one.


### Goal
Design and implement a Platform-as-a-Service offering on top of Kubernetes.

### Scope
1. **Understanding Kubernetes Concepts**
   - [ ] Infrastructure Components: Control Plane and Worker Nodes
   - [ ] How Pods are scheduled and executed on Worker Nodes
2. **SKE Cluster Creation:** Using the STACKIT Terraform Provider to provision an SKE cluster
   - [ ] Configure the STACKIT Terraform provider
   - [ ] Authenticate Terraform with the STACKIT platform
   - [ ] Provision an SKE (STACKIT Kubernetes Engine) cluster
   - [ ] Verify that the cluster is operational and accessible with `kubectl`
3. **PaaS Product Implementation (e.g., Managed Database):** Design and technical implementation of a simple PaaS service
   - [ ] Select and deploy a Kubernetes Operator for the chosen PaaS product.
   - [ ] Provision product components using Custom Resources (CRs).
   - [ ] Verify that the Operator successfully manages the deployed resources.
   - [ ] Document how to connect to and use the deployed PaaS service.
   - [ ] Demonstrate successful connectivity to the PaaS product.
4. **Deepening Kubernetes Knowledge:** Deepening knowledge of Custom Resource Definitions (CRDs) and the functioning of Operators (Reconciler Pattern)
   - [ ] Explain the purpose and structure of Custom Resource Definitions (CRDs).
   - [ ] Explain how Kubernetes Operators extend cluster functionality.
   - [ ] Describe the Reconciler Pattern and its role in resource management.

### Bonus
* **Automating the Deployment:** Introduction of a GitOps approach and CI/CD integration for automated provisioning of both SKE and the PaaS service.
   - [ ] Introduce a GitOps workflow for managing infrastructure and application deployments.
   - [ ] Integrate a CI/CD pipeline to automate the provisioning of the SKE cluster.
   - [ ] Automate the deployment of the PaaS service using the CI/CD pipeline.
   - [ ] Verify that infrastructure and application changes are applied automatically through the GitOps workflow.

---

## [Week 3 – Developing a Platform-as-a-Service (PaaS) Product: API Layer](./week-3/README.md)

APIs are the front door to your platform. Make them robust, secure, and developer-friendly.


### Goal
Expose the PaaS product through a clean, production-ready RESTful API to enable automated provisioning and seamless integration.


### Scope
1. **API Development for Product Instances:** Development of a RESTful API that provides operations for creating, deleting, and listing PaaS product instances, as well as retrieving associated connection and access data.
   - [ ] Develop a RESTful API for creating PaaS product instances.
   - [ ] Implement an endpoint for deleting PaaS product instances.
   - [ ] Implement an endpoint for listing existing PaaS product instances.
   - [ ] Implement an endpoint for retrieving connection and access information for a product instance.
   - [ ] **API Specification:** The functionality of the API must be fully documented using the OpenAPI specification.
2. **Unit Tests:** Implementation of unit tests for each endpoint of the developed API.
   - [ ] Implement unit tests for the product instance creation endpoint.
   - [ ] Implement unit tests for the product instance deletion endpoint.
   - [ ] Implement unit tests for the product instance listing endpoint.
   - [ ] Implement unit tests for the connection and access information endpoint.
3. **Docker Container Image:** Creating a Docker image for the API, uploading it to the STACKIT Container Registry, and deploying the API via SKE.
   - [ ] Create a Docker image for the API.
   - [ ] Push the Docker image to the STACKIT Container Registry.
   - [ ] Deploy the API to the SKE cluster.
   - [ ] Verify that the deployed API is accessible and functioning correctly.
4. **Understanding the Lifecycle Process (Create Flow):**
   - [ ] **Flowchart:** Visualization of the individual steps required to create a product instance.
   - [ ] Explain the sequence of operations performed during instance creation.
   - [ ] **REST Fundamentals:** Solidifying basic understanding of how a RESTful API operates.

### Bonus
* **Automated API Deployment:** Integration of the RESTful API deployment into the existing GitOps infrastructure.
   - [ ] Integrate the RESTful API deployment into the existing GitOps workflow.
   - [ ] Verify that API updates are deployed automatically through the GitOps pipeline.
* **Auto-Scaling and Performance Tests:**
   - [ ] **Horizontal Pod Autoscaler (HPA):** Configuration of a Kubernetes HPA for automatic scaling of the control plane RESTful API.
   - [ ] **Performance Tests:** Execution of performance tests for the RESTful API to verify the functionality of the HPA.
   - [ ] Configure a Horizontal Pod Autoscaler (HPA) for the RESTful API.
   - [ ] Define appropriate CPU and/or memory utilization thresholds for scaling.
   - [ ] Execute performance tests to generate load on the RESTful API.
   - [ ] Verify that the HPA scales the API pods as expected under load.
   - [ ] Document the performance test results and scaling behavior.
* **Update Functionality:** Implementation and testing of an API endpoint that allows updating access data and instance details.
   - [ ] Implement an API endpoint for updating product instance details.
   - [ ] Implement an API endpoint for updating connection and access data.
   - [ ] Test the update functionality to verify that changes are applied correctly.

---

## Week 4 – Extending the Platform with Advanced Features

Your platform isn’t complete until users can securely access and interact with it.

### Goal
Build a user-facing interface and expose the PaaS product securely via the web, ensuring a smooth and accessible developer experience.

### Scope
1. **Web UI Development:** Building a user-friendly web interface (using Vue.js or a similar framework) to interact with the PaaS product, covering all relevant API functions.
   - [ ] Develop a user-friendly web interface using Vue.js or a similar framework.
   - [ ] Implement functionality for creating PaaS product instances.
   - [ ] Implement functionality for listing existing PaaS product instances.
   - [ ] Implement functionality for deleting PaaS product instances.
   - [ ] Display connection and access information for product instances.
2. **Secure Communication:** Implementing secure communication between the UI and backend APIs using JWT (JSON Web Tokens) or a similar standard.
   - [ ] Implement authentication using JWT or a comparable standard.
   - [ ] Secure communication between the Web UI and backend APIs.
   - [ ] Verify that protected API endpoints require valid authentication.
3. **Ingress Controller Deployment:** Provisioning an Ingress controller on the SKE cluster.
   - [ ] Deploy an Ingress controller to the SKE cluster.
   - [ ] Configure Ingress resources for the Web UI and backend API.
   - [ ] Verify that external traffic is routed correctly.
4. **Production Publication:** Exposing the API and Web UI on the SKE cluster with SSL/TLS encryption and a free STACKIT subdomain URL.
   - [ ] Expose the Web UI through the Ingress controller.
   - [ ] Expose the backend API through the Ingress controller.
   - [ ] Configure SSL/TLS encryption for all external endpoints.
   - [ ] Publish the application using a free STACKIT subdomain.
   - [ ] Verify secure access over HTTPS.
5. **Architecture Diagram Update:** Adapting the architecture diagram to illustrate the exact traffic flow between the Web UI, Ingress, and backend APIs.
   - [ ] Update the architecture diagram to reflect the deployed solution.
   - [ ] Illustrate the traffic flow between the Web UI, Ingress, and backend APIs.
   - [ ] Include external access, TLS termination, and Kubernetes components in the diagram.

### Bonus
* **Automated UI Deployment:** Automating the Web UI deployment on the existing SKE via GitOps integration.
   - [ ] Integrate the Web UI deployment into the existing GitOps workflow.
   - [ ] Automate the deployment of the Web UI to the SKE cluster.
   - [ ] Verify that Web UI updates are deployed automatically through the GitOps pipeline.
* **End-to-End (E2E) Testing:** Implementation of E2E tests (e.g., via Cypress or Playwright) for both the Web UI and the RESTful API.
   - [ ] Implement E2E tests for the Web UI using Cypress, Playwright, or a similar framework.
   - [ ] Implement E2E tests for the RESTful API.
   - [ ] Validate the complete workflow from the Web UI to the backend API.
   - [ ] Integrate E2E tests into the CI/CD pipeline.
   - [ ] Verify that all E2E tests pass successfully.

---

## Week 5 – Extending the Platform with Advanced Features

If you can’t see it or audit it, you can’t run it at scale.

### Goal
Enable production-grade operations through observability and audit logging.

### Scope
1. **Internal Monitoring:** Gathering persistent telemetry data that the platform operator can use for real-time health and performance evaluation.
   - [ ] Deploy Prometheus and Grafana for monitoring Kubernetes cluster health and application performance.
   - [ ] Configure Prometheus to collect relevant infrastructure and application metrics.
   - [ ] Create Grafana dashboards for real-time monitoring and visualization.
   - [ ] Deploy Loki for centralized application log collection.
   - [ ] Configure log aggregation and analysis for deployed services.
2. **User-Centric Monitoring & Logging:** Persistently stored logs that users can retrieve for their specific instances via the UI or API.
   - [ ] Implementation of an Audit Logging system for recording user actions (creation, modification, or deletion of instances, access credentials, etc.).
   - [ ] Implement persistent storage for application logs.
   - [ ] Provide access to instance-specific logs through the Web UI or RESTful API.
   - [ ] Implement an Audit Logging system for tracking user actions.
   - [ ] Record instance lifecycle events such as creation, modification, and deletion.
   - [ ] Record access credential-related activities.

### Bonus
* **Development of a Golang SDK for the PaaS Product:**
   - [ ] Develop a native Go SDK for interacting with the PaaS product API.
   - [ ] Implement authentication and authorization support within the SDK.
   - [ ] Provide SDK usage documentation and examples.
   - [ ] Validate the SDK by demonstrating common API operations through the SDK.

---

## Final Result

By the end of this track, you will have:

* **Designed and operated a complete infrastructure stack** — from raw IaaS with OpenStack to a functional PaaS running on Kubernetes.
* **Provisioned and managed Kubernetes clusters** using Infrastructure as Code (IaC) with Terraform.
* **Built and deployed a Kubernetes-native PaaS product**, including Custom Resource Definitions (CRDs) and custom Operators.
* **Exposed the platform through a production-ready RESTful API** for end-to-end product lifecycle management.
* **Developed a secure, user-facing Web UI** that interfaces seamlessly with the API and runs on Kubernetes with proper Ingress and SSL.
* **Implemented observability and audit logging**, enabling platform-level monitoring, centralized logging, and absolute traceability for both operators and end users.

This is how real-world cloud platforms are designed, built, and operated from the ground up.