### **Project Goals and Motivation**

The primary objective is to establish a self-hosted, on-premise Generative AI infrastructure. This initiative is driven by two key concerns:

1.  **User Data Privacy:** Maintaining full control over student data by processing it in-house, eliminating reliance on third-party services.
2.  **Cost Predictability:** Mitigating the financial risk of future price increases common with managed cloud-based AI services.

The initial hardware budget is **$10,000 - $15,000**, intended to serve an active user base of **300 students**.

---

### **Proposed Technology and Data Strategy**

* **Core Model Selection:** We are considering **IBM Granite** as the foundational LLM. Its transparent training data and enterprise-friendly open-source license make it an ideal choice to address concerns about data provenance and potential copyright issues.
* **Data Management:** To ensure full copyright compliance, the data strategy requires that all materials used for training or reference are **explicitly purchased or licensed**. This content will be stored securely in an on-premise database.

---

### **Key Assumptions and Risks to Validate**

This project's feasibility depends on several critical assumptions that require validation:

* **Hardware Sufficiency:** We are assuming that a hardware setup within the specified **$10-15K budget** will provide the necessary computational power to run the selected LLMs effectively for our use case.
* **Network Bandwidth:** The plan assumes that a single server connected to our office's standard internet connection can reliably serve **300 concurrent students** with acceptable latency. This assumption presents a significant risk and must be tested to ensure service quality.