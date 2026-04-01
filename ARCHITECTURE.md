<!-- SPDX-FileCopyrightText: 2026 The templig contributors.
     SPDX-License-Identifier: MPL-2.0
-->

Architecture
============

*templig* is an easy-to-use configuration library. It hides the details of
loading files, expanding templating expressions and potential validation steps
from the programmer and user.

The loading of a configuration is explicated in the following picture:

```mermaid
graph TD
    subgraph Input_Layer [Input Layer]
        ConfigTemplate[Templated Config File: .yaml]
        EnvVars[Environment Variables]
        RefFiles[Referenced Files]
    end
        
    subgraph Core_Engine [templig Core]
        Loader[Template Loader]
        FuncMap[Additional Template Functions]
        Renderer[Go Template Engine]
        Parser[Config Parser]
        Validator[Config Validator]
    end

    subgraph Output_Layer [Output Layer]
        Config[Parsed Structure]
        SecretHiding[Secret Hiding]
        Stdout[Textual Output]
    end

    %% Data Flows
    ConfigTemplate --> Loader
    EnvVars --> Renderer
    RefFiles --> Renderer
    
    Renderer --> Parser
    Loader --> Renderer
    FuncMap --> Renderer
    Parser --> Validator
    
    Parser --> Config
    Config --> SecretHiding
    SecretHiding --> Stdout

    style Core_Engine fill:#f9f,stroke:#333,stroke-width:2px
```

The *templig* Core operates entirely in-memory and does not require network
access. All inputs (templates, environment variables) are processed through the
Go template engine before being validated against the expected structure. Should
a *templig* user decide to extend the core with functionality violating this
promise it is outside the scope of this document and project.
