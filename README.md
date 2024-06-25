# ws-mse-demo

prepare fragmented mp4 file media source buffer

```sh
#!/bin/bash

ffmpeg -i bunny_orig.mp4 -c:v libx264 -profile:v main -level 3.2 -pix_fmt yuv420p -b:v 2M -preset medium -tune zerolatency -flags +cgop+low_delay -movflags empty_moov+omit_tfhd_offset+frag_keyframe+default_base_moof+isml -an bunny_prepared.mp4

```
