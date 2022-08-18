---
title: Querying a public system
bookCollapseSection: false
weight: 1
---

# Cyber-dojo introduction

```shell
export KOSLI_API_TOKEN=<put your kosli API token here>
export KOSLI_OWNER=cyber-dojo
```

```shell
kosli env ls 
```
```shell
NAME      TYPE  LAST REPORT                LAST MODIFIED
aws-beta  ECS   2022-08-18T09:57:13+02:00  2022-08-18T09:57:13+02:00
aws-prod  ECS                              2022-08-12T15:12:44+02:00
```

We can find out all the running artifacts in the current `prod` environment snapshot:
```shell
kosli snapshot get prod                                    
```

```shell
COMMIT   ARTIFACT                                                                  PIPELINE                RUNNING_SINCE  REPLICAS
bbf94ef  Name: cyberdojo/web:bbf94ef                                               web                     3 months ago   3
         SHA256: 95f3d36bd1849b9caf4d014641bfe817384d8990477430e4287c92edb3a68762                                         
                                                                                                                          
f1c426f  Name: cyberdojo/runner:f1c426f                                            runner                  3 months ago   3
         SHA256: 9e490165d5e8f8a8260a7be37595328c3e3ff74252c1ca312ae64f3ebfad1636                                         
                                                                                                                          
68c5eb7  Name: cyberdojo/saver:68c5eb7                                             saver                   3 months ago   1
         SHA256: 8ba413cc804ecac73779925f0d97a021e7c13a0cbd8dd24eaaf27e833c3619e2                                         
                                                                                                                          
7e2c8b4  Name: cyberdojo/nginx:7e2c8b4                                             nginx                   3 months ago   1
         SHA256: ddb54f0c8b8c69143d617cc559c231e7389f1d7d8d875a5909aa3303e0397a3b                                         
                                                                                                                          
f0eeae4  Name: cyberdojo/languages-start-points:f0eeae4                            languages-start-points  3 months ago   1
         SHA256: a65c49270f831b89660603dee6d20b58a9b50febb72f90c3bbd08a18fa74ce69                                         
                                                                                                                          
c6d6a35  Name: cyberdojo/exercises-start-points:c6d6a35                            exercises-start-points  3 months ago   1
         SHA256: 76e4fef7e98a2248ac2705fee422d8e2e3ce1edb9109e8c0e2f7cb52c28c20c3                                         
                                                                                                                          
N/A      Name: grafana/grafana:6.2.4                                               N/A                     3 months ago   1
         SHA256: 773615f0ae8e783170e7084ac5173f22ead52b4bdbe7de1899725bcfc0555a97                                         
                                                                                                                          
916b024  Name: cyberdojo/shas:916b024                                              shas                    3 months ago   1
         SHA256: aadbfc30734b75369002c69e6232a47f45e202f19280747028b5d337a05645e5                                         
                                                                                                                          
a71729f  Name: cyberdojo/repler:a71729f                                            repler                  3 months ago   1
         SHA256: f740fb1897b22780583cc49c0b8460d4f6a4e56d603cbcc65f10a6132c2ff65a                                         
                                                                                                                          
7a05eb4  Name: cyberdojo/creator:7a05eb4                                           creator                 3 months ago   1
         SHA256: 12bfc09116a85d9fd427ff4542932880f1bac088a85b5ebb88cd74175c767807                                         
                                                                                                                          
ef2352f  Name: cyberdojo/custom-start-points:ef2352f                               custom-start-points     3 months ago   1
         SHA256: 1ea9ac6b3ad0e98b6b030e34cd30e330d09e0c9cc7eee7623ba06795364fd91e                                         
                                                                                                                          
f05a57c  Name: cyberdojo/differ:f05a57c                                            differ                  3 months ago   1
         SHA256: 83c8b5b2a65b7381a87eb43a92acddd2a1960bd8bc6164d0c38a5714d4675b7f                                         
                                                                                                                          
0aed98e  Name: cyberdojo/dashboard:0aed98e                                         dashboard               3 months ago   1
         SHA256: a3b2190b68c7c2702b2358477629617a12c820fe02e3da32c516b824b9029497                                         
```

You can find what has changed in `prod` compared to previous snapshot:

```shell
kosli env diff prod~1 prod
```
```shell
- Name:   cyberdojo/web:bbf94ef
  Sha256: 95f3d36bd1849b9caf4d014641bfe817384d8990477430e4287c92edb3a68762
  Pipeline: web
  Commit: https://github.com/cyber-dojo/web/commit/bbf94ef0cf8b9be9357870b6ff6e08d90d086905
  Pods:   [web-59hpf web-7j97f web-p9tbr web-xs272]
  Started: 03 Jun 22 23:02 CEST • 3 months ago
- Name:   cyberdojo/runner:f1c426f
  Sha256: 9e490165d5e8f8a8260a7be37595328c3e3ff74252c1ca312ae64f3ebfad1636
  Pipeline: runner
  Commit: https://github.com/cyber-dojo/runner/commit/f1c426f24b771b2d54bbd9752927a63a71b58cf4
  Pods:   [runner-9xbw6 runner-f9gch runner-k9q8g runner-p24hn]
  Started: 03 Jun 22 23:02 CEST • 3 months ago

+ Name:   cyberdojo/web:bbf94ef
  Sha256: 95f3d36bd1849b9caf4d014641bfe817384d8990477430e4287c92edb3a68762
  Pipeline: web
  Commit: https://github.com/cyber-dojo/web/commit/bbf94ef0cf8b9be9357870b6ff6e08d90d086905
  Pods:   [web-7j97f web-p9tbr web-xs272]
  Started: 03 Jun 22 23:02 CEST • 3 months ago
+ Name:   cyberdojo/runner:f1c426f
  Sha256: 9e490165d5e8f8a8260a7be37595328c3e3ff74252c1ca312ae64f3ebfad1636
  Pipeline: runner
  Commit: https://github.com/cyber-dojo/runner/commit/f1c426f24b771b2d54bbd9752927a63a71b58cf4
  Pods:   [runner-9xbw6 runner-k9q8g runner-p24hn]
  Started: 03 Jun 22 23:02 CEST • 3 months ago
```
#TODO: use a diff to find something interesting (e.g. broken prod or bug)
