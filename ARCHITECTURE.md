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
        ConfigFile[/Templated Config File: .yaml/]
        ConfigTemplate[/Templated Config YAML/]
        EnvVars[/Environment Variables/]
        RefFiles[/Referenced Files/]
        ConfigTypes[/Config Types/]
    end
        
    subgraph Core_Engine [templig Core]
        FileLoader[File Loader]
        TemplateLoader[Template Loader]
        FuncMap[Additional Template Functions]
        Renderer[Go Template Engine]
        Parser[Config Parser]
        HasValidator{Has Validator}
        Validator[Config Validator]

        FileLoader     -- "Reader" --> TemplateLoader
        Renderer       -- "YAML"   --> Parser
        TemplateLoader             --> Renderer
        FuncMap                    --> Renderer
        Parser                     --> HasValidator
        HasValidator   -- "yes"    --> Validator
    end

    subgraph Output_Layer [Output Layer]
        Config[/Parsed Structure/]
        WithSecretHiding{WithSecretHiding}
        SecretHiding[Secret Hiding]
        Stdout[/Textual Output/]
        
        Config                             --> WithSecretHiding
        WithSecretHiding -- "yes\n\nYAML"  --> SecretHiding
        WithSecretHiding -- "no\n\nWriter" --> Stdout
        SecretHiding     -- "Writer"       --> Stdout
    end

    %% Input --> Core
    ConfigFile                 --> FileLoader
    ConfigTemplate -- "Reader" --> TemplateLoader
    EnvVars                    --> Renderer
    RefFiles                   --> Renderer
    ConfigTypes                --> Parser
   
    %% Core --> Output
    HasValidator   --"no" --> Config
    Validator             --> Config

    style Core_Engine  fill:#f9f,stroke:#333,stroke-width:2px
    style Validator    stroke-dasharray: 5 5
    style SecretHiding stroke-dasharray: 5 5
```

*templig* utilizes Go generics to ensure type safety. The Config Types provided
at the input layer act as a blueprint for the Config Parser. This ensures that
the rendered output is not only syntactically correct but also semantically
compliant with the user's expected data structure.

The *templig* Core operates entirely in-memory and does not require network
access. All inputs (templates, environment variables) are processed through the
Go template engine before being validated against the expected structure. Should
a *templig* user decide to extend the core with functionality violating this
promise it is outside the scope of this document and project.
