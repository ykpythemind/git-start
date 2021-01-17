# git-start

## build

```
$ go build -o /usr/local/bin/git-start ./cmd/start
```

`git start` で実行可能になる

## 解説

- [課題] 以下の手順を毎回行っているが毎回ブラウザやエディタの切り替えでコンテキストスイッチが発生してウーンとなっていた。
  - GitHubのissue内容を見る
  - 適切にブランチ名をつける
  - 実装
  - git push
  - PRを作成する
- どうせならCLIから完結するようにした
  - ![2021-01-17 22 35 46](https://user-images.githubusercontent.com/22209702/104844845-c806b480-5915-11eb-870f-2505f455db68.gif)
  - `git start [issue num] | [issue url]`
    - GitHubのAPIからイシュー内容を取得し、エディタに展開する
    - ブランチ名, 後で使うPRタイトル（実装内容）を入力してエディタを終了するとブランチが作成される
      - git commit時にエディタが起動してコミットメッセージ書くUXに近い形
  - 実装する
  - `git start pr`
    - ブランチをpushし、ブラウザを起動する (PRタイトルはブランチ作成時に決めたものがデフォで入力された状態）
- todo
  - 俗に言う生煮えPR機能？
  - checkoutした瞬間にempty commitしてPR作成しに行くとか? 
  - GitLab等対応
  - git config localに issue repositoryを設定する機能
  - 複数remoteがある時の諸々
- 実装済
  - git remoteから issue urlを推測するのでイシュー番号だけでもなんとかなる機能

## token

- プライベートレポジトリのissueを取得するために personal access tokenが必要
- 環境変数 GITHUB_TOKEN ないしは ~/.git-start-token に記述する
