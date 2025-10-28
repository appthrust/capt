## Cluster API v1beta2 変更点・移行ガイド（調査レポート）

本書は Cluster API（CAPI）における v1beta2 の変更点を整理し、移行観点での要点をまとめたものです。対象読者は CAPI を用いたコントローラー/プロバイダー開発者および運用者です。

### 1. 位置づけとサポートポリシー

- 契約（Contract）: CAPI の管理クラスターは「契約」バージョンにより互換性を管理します。v1beta2 契約が現行で、管理クラスター内の全プロバイダーは同一契約で揃える必要があります（`clusterctl` が初期化/アップグレード時に検査）。
- v1beta1 の扱い: v1.11 で非推奨、v1.14（2025-08-26 予定）で提供終了（EOL）。v1beta2 と v1beta1 は一時的に共存可能だが将来的に終了予定。
- Kubernetes サポート範囲（例）: 各 CAPI マイナーはおおむね 管理 N..N-3 / ワークロード N..N-5 のマイナー範囲をサポート。例: CAPI v1.7 は管理 1.26–1.29 / ワークロード 1.24–1.29。
- ダウングレード不可: CAPI はダウングレードをサポートしません。n−3 より古い版からの直接アップグレードも非推奨。

参考: 公式 Version/Contract ポリシー（サポート期間・互換性・n/n-3/n-5 マトリクス）
- https://main.cluster-api.sigs.k8s.io/reference/versions

### 2. API スキーマの主な変更（v1beta1 → v1beta2）

以下はコア API（`cluster.x-k8s.io/*`）の代表的差分です。詳細は各プロバイダー/実装に依存する部分もあるため、型定義と CRD をあわせて確認してください。

#### 2.1 ObjectReference 系の `ref` から `templateRef` への統一

- 変更例:
  - `spec.infrastructure.ref` → `spec.infrastructure.templateRef`
  - `spec.controlPlane.ref` → `spec.controlPlane.templateRef`
  - `spec.controlPlane.machineInfrastructure.ref` → `spec.controlPlane.machineInfrastructure.templateRef`
  - `spec.workers.machineDeployments[].template.bootstrap.ref` → `spec.workers.machineDeployments[].bootstrap.templateRef`
  - `spec.workers.machineDeployments[].template.infrastructure.ref` → `spec.workers.machineDeployments[].infrastructure.templateRef`
  - `spec.workers.machinePool[].template.bootstrap.ref` → `spec.workers.machinePool[].bootstrap.templateRef`
  - `spec.workers.machinePool[].template.infrastructure.ref` → `spec.workers.machinePool[].infrastructure.templateRef`

- 影響:
  - `*.ref` から `namespace`, `uid`, `resourceVersion`, `fieldPath` が削除。
  - 参照の意味合いが「テンプレート参照」に揃えられ、可読性と一貫性が向上。

#### 2.2 HealthCheck 構造の再編成

- 変更例:
  - `spec.controlPlane.machineHealthCheck` → `spec.controlPlane.healthCheck`
  - `spec.workers.machineDeployments[].machineHealthCheck` → `spec.workers.machineDeployments[].healthCheck`
  - `spec.workers.machineDeployments[].healthCheck.remediation.maxInFlight` が従来の `strategy` 直下から HealthCheck 側に集約

- 影響:
  - ヘルスチェック定義と修復（remediation）の責務が明確化。
  - コントロールプレーン/ワーカーでの指定方法が統一。

#### 2.3 Duration フィールドの秒数表現・型見直し

- 変更例:
  - `nodeStartupTimeout` → `nodeStartupTimeoutSeconds`（型: `*int32`）
  - `unhealthyNodeConditions[].timeout` → `unhealthyNodeConditions[].timeoutSeconds`（型: `*int32`）

- 影響:
  - Kubernetes API ガイドラインに従い、単位の明確化と型のシンプル化。

#### 2.4 Remediation Template 参照の型変更

- 変更例:
  - `remediation.templateRef` の型が `corev1.ObjectReference` → `MachineHealthCheckRemediationTemplateReference`
  - 伴って `namespace`, `uid`, `resourceVersion`, `fieldPath` が削除

#### 2.5 その他の代表的変更

- `spec.controlPlane.healthCheck.remediation.triggerIf.unhealthyInRange`: `*string` → `string`
- `spec.workers.machineDeployments[].template.metadata` → `spec.workers.machineDeployments[].metadata`（不要なネスト解消）
- `spec.workers.machinePools[].template.metadata` → `spec.workers.machinePools[].metadata`
- `spec.infrastructureNamingStrategy` → `spec.infrastructure.naming`（`InfrastructureClassNamingSpec`）
- `spec.controlPlane.namingStrategy.template` → `spec.controlPlane.naming`（`ControlPlaneClassNamingSpec`）
- `spec.workers.machineDeployments[].failureDomain`: `*string` → `string`
- `spec.workers.machineDeployments[].deletion.order`: `*string` → `MachineSetDeletionOrder`
- `spec.workers.machineDeployments[].rollout` の導入（従来の `strategy` を内包）
- `spec.workers.machineDeployments[].namingStrategy` → `spec.workers.machineDeployments[].naming`（`MachineDeploymentClassNamingSpec`）

### 3. ClusterClass / Topology / Runtime Extensions

- ClusterClass/Topology: v1beta2 に合わせ、上記の `templateRef` 化や HealthCheck の統一が反映。Topology 変数/パッチ運用は従来どおりだが、フィールド名・型変更の伝搬に留意。
- Runtime Extensions（フック）: 既存の拡張ポイントは継続。v1beta2 での大きな追加/削除は確認されていないが、型変更に伴うフックの受け口の更新に注意。

（注）上記の細部は利用プロバイダー/バージョンに依存。各プロバイダーの v1beta2 契約対応版リリースノートを確認すること。

### 4. 移行ガイド（実務要点）

1. 管理クラスター/プロバイダー整合
   - `clusterctl upgrade` を用い、管理クラスター内の「全」プロバイダーを v1beta2 契約対応へ統一。
   - v1beta1 との一時共存は移行のための猶予。早期に v1beta2 へ揃える。

2. CRD とストレージバージョン
   - CRD を v1beta2 へ更新し、ストレージバージョンを v1beta2 に切替。
   - 変換 Webhook がある前提で、既存 CR を v1beta2 へ自動変換。変換完了後に旧バージョン提供を停止。

3. マニフェスト/コードの更新
   - `apiVersion: cluster.x-k8s.io/v1beta2` へ更新。
   - `*.ref` → `*.templateRef` など、セマンティクス含めて差分を一括反映。
   - HealthCheck/Remediation/Duration（Seconds化・int32化）等の型差分に追従。

4. 検証と段階適用
   - ダウングレード不可のため、テスト環境でフル検証→段階的リリース。
   - Kubernetes バージョン窓口（管理 N..N-3、ワークロード N..N-5）を満たすこと。

### 5. 互換性/既知の注意点

- v1beta1 は v1.14 で EOL（2025-08-26 予定）。それ以前に v1beta2 へ移行完了する計画が必要。
- v1alpha3/v1alpha4 は提供終了済み。古い CR/CRD が残るとガーベジコレクション挙動に影響する可能性があるため、完全変換後に整理する。
- K8s 本体の非推奨 API 削除（例: 1.26 で flowcontrol v1beta1 等）も同時に考慮し、管理/ワークロードの周辺マニフェストを最新化。

### 6. 参考リンク（一次情報）

- Version/Contract サポート方針（公式）
  - https://main.cluster-api.sigs.k8s.io/reference/versions
- v1.10 → v1.11 の移行ガイド（型差分の代表例がまとまっている）
  - https://cluster-api.sigs.k8s.io/developer/providers/migrations/v1.10-to-v1.11

（補足）各インフラ/ブートストラップ/コントロールプレーンプロバイダー（CAPA/CAPZ/CAPV/CAPG 等）の v1beta2 対応状況・追加変更は、それぞれのリリースノートをご確認ください。


