package eks_blueprints_addons

// GenerateHCL converts the EKSBlueprintsAddonsConfig to Terraform HCL
// func (c *EKSBlueprintsAddonsConfig) GenerateHCL() (string, error) {
// 	f := hclwrite.NewEmptyFile()
// 	rootBody := f.Body()

// 	// Add module block for EKS Blueprints Addons
// 	moduleBlock := rootBody.AppendNewBlock("module", []string{"eks_blueprints_addons"})
// 	moduleBody := moduleBlock.Body()

// 	moduleBody.SetAttributeValue("source", cty.StringVal("aws-ia/eks-blueprints-addons/aws"))
// 	moduleBody.SetAttributeValue("version", cty.StringVal("~> 1.16"))

// 	// Add cluster configuration
// 	moduleBody.SetAttributeValue("cluster_name", cty.StringVal(c.ClusterName))
// 	moduleBody.SetAttributeValue("cluster_endpoint", cty.StringVal(c.ClusterEndpoint))
// 	moduleBody.SetAttributeValue("cluster_version", cty.StringVal(c.ClusterVersion))
// 	moduleBody.SetAttributeValue("oidc_provider_arn", cty.StringVal(c.OIDCProviderARN))

// 	// Add EKS Addons
// 	eksAddonsBlock := moduleBody.AppendNewBlock("eks_addons", nil)
// 	eksAddonsBody := eksAddonsBlock.Body()
// 	for addonName, addon := range c.EKSAddons {
// 		addonBlock := eksAddonsBody.AppendNewBlock(addonName, nil)
// 		addonBody := addonBlock.Body()
// 		if addon.ConfigurationValues != "" {
// 			tokens, err := generateTokensForExpression(addon.ConfigurationValues)
// 			if err != nil {
// 				return "", fmt.Errorf("failed to generate tokens for addon %s: %w", addonName, err)
// 			}
// 			addonBody.SetAttributeRaw("configuration_values", tokens)
// 		}
// 	}

// 	moduleBody.SetAttributeValue("enable_karpenter", cty.BoolVal(c.EnableKarpenter))

// 	// Add Karpenter configuration
// 	karpenterBlock := moduleBody.AppendNewBlock("karpenter", nil)
// 	karpenterBody := karpenterBlock.Body()
// 	helmConfigBlock := karpenterBody.AppendNewBlock("helm_config", nil)
// 	helmConfigBody := helmConfigBlock.Body()
// 	helmConfigBody.SetAttributeValue("cacheDir", cty.StringVal(c.Karpenter.HelmConfig.CacheDir))

// 	// Add Karpenter Node configuration
// 	karpenterNodeBlock := moduleBody.AppendNewBlock("karpenter_node", nil)
// 	karpenterNodeBody := karpenterNodeBlock.Body()
// 	karpenterNodeBody.SetAttributeValue("iam_role_use_name_prefix", cty.BoolVal(c.KarpenterNode.IAMRoleUseNamePrefix))

// 	// Add tags
// 	if len(c.Tags) > 0 {
// 		tagsObj := make(map[string]cty.Value)
// 		for k, v := range c.Tags {
// 			tagsObj[k] = cty.StringVal(v)
// 		}
// 		moduleBody.SetAttributeValue("tags", cty.ObjectVal(tagsObj))
// 	}

// 	// Format the generated HCL
// 	var buf bytes.Buffer
// 	_, err := f.WriteTo(&buf)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate HCL: %w", err)
// 	}

// 	return buf.String(), nil
// }

// func generateTokensForExpression(expr string) (hclwrite.Tokens, error) {
// 	expr = strings.TrimSpace(expr)
// 	if expr == "" {
// 		return nil, fmt.Errorf("empty expression")
// 	}

// 	// Check if the expression is a simple string (no spaces or special characters)
// 	if !strings.ContainsAny(expr, " \t\n\r{}[]") {
// 		return hclwrite.Tokens{
// 			&hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(expr)},
// 		}, nil
// 	}

// 	// If it's not a simple string, treat it as a complex expression
// 	tokens := hclwrite.Tokens{
// 		&hclwrite.Token{Type: hclsyntax.TokenOQuote, Bytes: []byte{'"'}},
// 	}

// 	// Escape any double quotes in the expression
// 	expr = strings.ReplaceAll(expr, `"`, `\"`)

// 	tokens = append(tokens, &hclwrite.Token{
// 		Type:  hclsyntax.TokenQuotedLit,
// 		Bytes: []byte(expr),
// 	})

// 	tokens = append(tokens, &hclwrite.Token{
// 		Type:  hclsyntax.TokenCQuote,
// 		Bytes: []byte{'"'},
// 	})

// 	return tokens, nil
// }
