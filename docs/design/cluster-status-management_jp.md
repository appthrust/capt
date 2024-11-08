# クラスターステータス管理の設計

## 概要

このドキュメントでは、CAPTプロバイダーにおけるクラスターステータスの管理に関する設計と実装の詳細について説明します。

## 背景

ClusterAPIは、インフラストラクチャプロバイダーに対して、自身のリソース（CAPTCluster、CAPTControlPlane）とコアClusterリソースの両方のステータスを管理することを要求します。これには以下が含まれます：

1. インフラストラクチャの準備状態
2. コントロールプレーンの準備状態
3. クラスターフェーズの遷移
4. ステータス条件とエラー処理

## ClusterAPIからの要件

### インフラストラクチャプロバイダーの要件

1. **オーナー参照**
   - インフラストラクチャプロバイダーはインフラストラクチャオブジェクトにオーナー参照を設定する必要がある
   - Clusterコントローラーは`Cluster.spec.infrastructureRef`で参照されるインフラストラクチャオブジェクトにオーナー参照を設定する

2. **ステータス管理**
   - 必須フィールドを持つステータスオブジェクトを提供する必要がある：
     - `ready` - インフラストラクチャが準備完了かどうかを示すブール値
     - `controlPlaneEndpoint` - クラスターのAPIサーバーに接続するためのエンドポイント

3. **オプションのステータスフィールド**
   - `failureReason` - 致命的なエラーが発生した理由を説明する文字列
   - `failureMessage` - 詳細なエラーメッセージ
   - `failureDomains` - マシン配置のための障害ドメインのマップ

### クラスターフェーズ管理

クラスターは以下のフェーズを経ます：
1. `Provisioning` - インフラストラクチャとコントロールプレーンの準備中の初期状態
2. `Provisioned` - インフラストラクチャとコントロールプレーンの準備完了
3. `Running` - クラスターが完全に稼働中

フェーズの遷移は以下によって決定されます：
- CAPTClusterからの`InfrastructureReady`ステータス
- CAPTControlPlaneからの`ControlPlaneReady`ステータス

## 実装の詳細

### CAPTClusterコントローラー

1. **ステータス更新**
   ```go
   func (r *CAPTClusterReconciler) updateStatus(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
       // まずCAPTClusterのステータスを更新
       if err := r.Status().Update(ctx, captCluster); err != nil {
           return err
       }

       // Clusterが存在する場合はそのステータスも更新
       if cluster != nil {
           patch := client.MergeFrom(cluster.DeepCopy())
           cluster.Status.InfrastructureReady = captCluster.Status.Ready
           
           // インフラストラクチャとコントロールプレーンの両方が準備完了の場合はフェーズを更新
           if cluster.Status.InfrastructureReady && cluster.Status.ControlPlaneReady {
               cluster.Status.Phase = string(clusterv1.ClusterPhaseProvisioned)
           }

           if err := r.Status().Patch(ctx, cluster, patch); err != nil {
               return err
           }
       }

       return nil
   }
   ```

2. **オーナー参照**
   ```go
   func (r *CAPTClusterReconciler) setOwnerReference(ctx context.Context, captCluster *infrastructurev1beta1.CAPTCluster, cluster *clusterv1.Cluster) error {
       if cluster == nil {
           return nil
       }

       // オーナー参照が既に設定されているか確認
       for _, ref := range captCluster.OwnerReferences {
           if ref.Kind == "Cluster" && ref.APIVersion == clusterv1.GroupVersion.String() {
               return nil
           }
       }

       return controllerutil.SetControllerReference(cluster, captCluster, r.Scheme)
   }
   ```

### CAPTControlPlaneコントローラー

CAPTClusterと同様に、CAPTControlPlaneコントローラーは：
1. 自身のステータスを更新
2. ClusterのControlPlaneReadyステータスを更新
3. コントロールプレーンのエンドポイント情報を提供

## 重要な考慮事項

1. **ステータス更新の順序**
   - プロバイダーのステータス（CAPTCluster/CAPTControlPlane）を常にClusterステータスの更新前に更新する
   - 競合を避けるためにClusterステータスの更新にはサーバーサイドアプライ（パッチ）を使用する

2. **フェーズ遷移**
   - フェーズ遷移は決定論的で、明確な条件に基づいている必要がある
   - Provisionedフェーズに移行する前に、インフラストラクチャとコントロールプレーンの両方が準備完了である必要がある

3. **エラー処理**
   - ステータス条件を通じて詳細なエラー情報を伝播する
   - ClusterAPIから適切なエラータイプを使用する（例：`ClusterStatusError`）

4. **調整**
   - 調整ループを防ぐために不要なステータス更新を避ける
   - 可能な場合は更新の代わりにパッチ操作を使用する
   - 変更検出にgeneration/observedGenerationの使用を検討する

## 将来の改善点

1. **ステータス条件**
   - より良い観測性のために、より詳細なステータス条件を実装
   - インフラストラクチャプロビジョニングの異なるステージのための条件を追加

2. **障害ドメインのサポート**
   - 障害ドメインとしてAWSアベイラビリティゾーンのサポートを追加
   - 障害ドメインを考慮したマシン配置を実装

3. **フェーズ管理**
   - より細かいフェーズ遷移のサポートを追加
   - 劣化状態のより良い処理を実装

4. **ステータス更新**
   - ステータス更新の頻度を最適化
   - ステータス更新の競合解決を改善
