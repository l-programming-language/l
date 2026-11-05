package contracts

// VerifySourceAdjective caches the scoped logical theorem only. The frontend
// must resolve and project the source first. Current provenance and lowering
// obligations are always refreshed, even when semantic projection is unchanged.
func (session *Session) VerifySourceAdjective(projection SourceAdjectiveProjection, declaration SourceTransformerDeclaration, options AnalysisOptions) (SourceAdjectiveCertificate, error) {
	options = options.logical()
	logicalProjection := projection
	logicalProjection.Dependencies = nil
	logicalProjection.LoweringObligations = nil
	if logicalProjection.Environment != "" {
		logicalProjection.Environment = "current-binding-revalidated"
	}
	key := logicalCacheKey{Schema: SchemaVersion, Meaning: projection.Meaning, Specification: StableID("source-declaration", declaration), Substitutions: StableID("source-semantics", logicalProjection), FacetSchemas: "source-structure-only/1", RoleInterfaces: StableID("source-nominal", projection.TargetNominal), Rules: SourceProjectionVersion + ":" + inferencePolicy, TargetModel: "source_elaboration", Dependencies: map[string]string{"analysis-options": StableID("options", options)}}
	certificate, err := cachedResult(session, key, func() (SourceAdjectiveCertificate, error) {
		return VerifySourceAdjective(projection, declaration, options)
	})
	if err != nil {
		return certificate, err
	}
	certificate.ProjectionFingerprint = StableID("source-projection", projection)
	certificate.Environment = projection.Environment
	certificate.Dependencies = projection.Dependencies
	certificate.LoweringObligations = append([]Obligation(nil), projection.LoweringObligations...)
	return certificate, nil
}

// VerifySourceAdjectiveInstance caches concrete structural applicability only.
// Its key includes current source/native dependencies, exact before/after
// attachment and current trust inputs. Native lowering obligations remain open.
func (session *Session) VerifySourceAdjectiveInstance(projection SourceAdjectiveProjection, certificate SourceAdjectiveCertificate, declaration SourceTransformerDeclaration, instance SourceAdjectiveInstance, options AnalysisOptions) (SourceAdjectiveInstantiation, error) {
	verified, err := session.VerifySourceAdjective(projection, declaration, options)
	if err != nil {
		return SourceAdjectiveInstantiation{}, err
	}
	logical := verified
	logical.Environment = ""
	logical.Dependencies = nil
	logical.LoweringObligations = nil
	logical.ProjectionFingerprint = ""
	key := implementationCacheKey{Schema: SchemaVersion, Logical: StableID("source-logical", logical), Implementation: StableID("source-projection", projection), Target: StableID("source-instance", instance), TrustPolicy: StableID("approved", options.ApprovedAssumptions), Witnesses: StableID("witnesses", options.NativeWitnesses), Dependencies: projection.Dependencies}
	// A supplied certificate is only a dependency check, never cached proof.
	key.Dependencies = make(map[string]string, len(projection.Dependencies)+1)
	for name, value := range projection.Dependencies {
		key.Dependencies[name] = value
	}
	key.Dependencies["native-applications"] = StableID("native-applications", options.NativeApplications)
	key.Dependencies["supplied-certificate"] = StableID("source-certificate", certificate)
	return cachedResult(session, key, func() (SourceAdjectiveInstantiation, error) {
		return instantiateVerifiedSourceAdjective(projection, verified, certificate, declaration, instance), nil
	})
}
