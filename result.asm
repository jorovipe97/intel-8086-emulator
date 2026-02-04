; listing_0045_challenge_register_movs
bits 16
mov ax, 8738
; ax: 0 -> 8738
mov bx, 17476
; bx: 0 -> 17476
mov cx, 26214
; cx: 0 -> 26214
mov dx, -30584
; dx: 0 -> 34952
mov word ss, ax
; ss: 0 -> 8738
mov word ds, bx
; ds: 0 -> 17476
mov word es, cx
; es: 0 -> 26214
mov al, 17
; ax: 8738 -> 8721
mov bh, 51
; bx: 17476 -> 13124
mov cl, 85
; cx: 26214 -> 26197
mov dh, 119
; dx: 34952 -> 30600
mov ah, bl
; ax: 8721 -> 17425
mov cl, dh
; cx: 26197 -> 26231
mov word ss, ax
; ss: 8738 -> 17425
mov word ds, bx
; ds: 17476 -> 13124
mov word es, cx
; es: 26214 -> 26231
mov sp, ss
; sp: 0 -> 17425
mov bp, ds
; bp: 0 -> 13124
mov si, es
; si: 0 -> 26231
mov di, dx
; di: 0 -> 30600
