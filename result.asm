; listing_0044_register_movs
bits 16
mov ax, 1
; ax: 0 -> 1
mov bx, 2
; bx: 0 -> 2
mov cx, 3
; cx: 0 -> 3
mov dx, 4
; dx: 0 -> 4
mov sp, ax
; sp: 0 -> 1
mov bp, bx
; bp: 0 -> 2
mov si, cx
; si: 0 -> 3
mov di, dx
; di: 0 -> 4
mov dx, sp
; dx: 4 -> 1
mov cx, bp
; cx: 3 -> 2
mov bx, si
; bx: 2 -> 3
mov ax, di
; ax: 1 -> 4
