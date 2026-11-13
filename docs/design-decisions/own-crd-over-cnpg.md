# 2. A Narrow CRD of My Own Wraps CloudNativePG

The API and UI never touch CNPG objects directly.

**Why:** a five-field spec and a normalised `phase` give a stable contract. CNPG upgrades, or a different engine, stay contained inside the operator.

**Trade-off:** it's one more component to build, test and run, with two reconciliation layers. The operator's CNPG API version must also be kept in sync with the CNPG version installed in the cluster.
