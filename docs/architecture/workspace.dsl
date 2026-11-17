# C4 model of the DBaaS platform. Both diagrams on the Architecture page are
# views of this one model: an arrow declared here appears on every level.
#
#   make diagrams          render the views to docs/architecture/diagrams/*.svg
#   make diagrams-preview  edit with live preview at http://localhost:8080

workspace "DBaaS Platform" "Self-service PostgreSQL on Kubernetes." {

    !identifiers flat

    model {
        developer = person "Developer" "Wants a PostgreSQL database without running it themselves."

        app = softwareSystem "Developer's application" "The developer's own service that stores its data in the database." {
            tags "External"
        }
        zitadel = softwareSystem "ZITADEL" "Identity provider: login and roles." {
            tags "External"
        }

        dbaas = softwareSystem "DBaaS Platform" "Creates and manages PostgreSQL databases on Kubernetes." {
            ui = container "Web UI" "Create, view and delete databases; run SQL." "Vue 3"
            api = container "REST API" "Checks who the user is and records their requests." "Go, Gin"
            operator = container "dbaas-operator" "Turns each request into a running database." "Go, Kubebuilder"
            cnpg = container "CloudNativePG" "Runs PostgreSQL: replication, failover, passwords." "Kubernetes operator"
            postgres = container "PostgreSQL databases" "One database cluster per request." "PostgreSQL" {
                tags "Database"
            }
        }

        developer -> ui "Manages databases" "HTTPS"
        app -> postgres "Reads and writes data" "SQL over TCP"

        ui -> api "Sends requests" "JSON/HTTPS"
        ui -> zitadel "Logs the user in" "OIDC"
        api -> zitadel "Validates the user's token" "OAuth2"
        api -> operator "Asks for a database" "PostgreSQL resource in Kubernetes"
        operator -> cnpg "Creates a database cluster" "CNPG Cluster resource"
        cnpg -> postgres "Runs and looks after" "Pods and volumes"
    }

    views {
        systemContext dbaas "context" "Level 1: who uses the platform, and what it depends on." {
            include *
            autolayout lr
        }

        container dbaas "containers" "Level 2: the parts inside the platform." {
            include *
            autolayout lr
        }

        properties {
            # Tags give external systems their grey. The legend would list every
            # tag combination, so hack/docs/diagrams.sh strips it.
            "c4plantuml.tags" "true"
        }

        styles {
            element "Person" {
                shape Person
                background #08427b
                color #ffffff
            }
            element "Software System" {
                background #1168bd
                color #ffffff
            }
            element "External" {
                background #999999
                color #ffffff
            }
            element "Container" {
                background #438dd5
                color #ffffff
            }
            element "Database" {
                shape Cylinder
            }
            relationship "Relationship" {
                style solid
            }
        }
    }

}
