---
title: "Naming Conventions"
bookCollapseSection: true
weight: 100
---

# Naming Conventions

Clear and consistent naming makes it easy for everyone to understand what each item in Kosli represents. Good names help you route attestations correctly and quickly find what you need.

Use these conventions for:
- **Flows** and **Trails**
- **Attestation Types**
- **Environments**

## General Guidelines

The general guidelines should be considered best practices for all naming conventions in Kosli. You can adapt them to fit your organization’s needs, but consistency is key. All of our proposed conventions follow these general guidelines:

**Structure**: `<element 1>` `<delimiter>` `<element 2>` `<delimiter>`...`<element N>`


1. **Choose Delimiter**: Choose a delimiter that works for your and stick with it consistently.</br>
For example hyphen `-`, underscore `_`, tilde `~` or dot `.`. </br>
Avoid mixing delimiters within the same naming scheme.
2. **Choose case style for elements**: Choose a meaningful case style across elements (e.g., PascalCase, camelCase, snake_case) and use it consistently. Avoid spaces and clashes with delimiters.
3. **Keep it concise**: Shorter names are easier to read and remember. Aim for concise but descriptive names.
4. **Avoid special characters**: Stick to alphanumeric characters and underscores/hyphens

{{% hint warning %}}
Be aware of using underscore `_` as the delimiter, as that conflicts with snake_case for elements.
{{% /hint %}}

{{% hint info %}}
The rest of this document uses hyphen `-` as the delimiter in examples, but you can choose any delimiter that fits your needs.
{{% /hint %}}

### Regular Expression

To help enforce these conventions programmatically, here are sample regular expressions you can use based on your chosen case style.

Adjust the regex if you choose a different delimiter.

{{< tabs "naming-regex" >}}
{{< tab "snake_case" >}}

**Example**: `element_one`-`element_two`-`element_three`

```bash
^[a-z][a-z0-9_]*(-[a-z][a-z0-9_]*)*$
```

{{< /tab >}}
{{< tab "camelCase" >}}

**Example**: `elementOne`-`elementTwo`-`elementThree`

```bash
^[a-z][a-zA-Z0-9]*(-[a-z][a-zA-Z0-9]*)*$
```

{{< /tab >}}
{{< tab "PascalCase" >}}

**Example**: `ElementOne`-`ElementTwo`-`ElementThree`

```bash
^[A-Z][a-zA-Z0-9]*(-[A-Z][a-zA-Z0-9]*)*$
```
{{< /tab >}}
{{< /tabs >}}


{{% hint info %}}
If you want a specific length limit (e.g., max 50 characters), you can add a lookahead at the start of the regex:

```bash
^(?=.{1,50}$) # + rest of the regex
```

You can use online regex testers like [regex101](https://regex101.com/) to validate and test these expressions.

{{% /hint %}}
