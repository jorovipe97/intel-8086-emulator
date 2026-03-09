; listing_0048_ip_register
bits 16
mov cx, 200
; cx: 0 -> 200
mov bx, cx
; bx: 0 -> 200
add cx, 1000
; cx: 200 -> 1200
mov bx, 2000
; bx: 200 -> 2000
sub cx, bx
; SF set
; CF set
; cx: 1200 -> 64736
