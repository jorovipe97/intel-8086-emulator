; listing_0046_add_sub_cmp
bits 16
mov bx, -4093
; bx: 0 -> 61443
mov cx, 3841
; cx: 0 -> 3841
sub bx, cx
; SF set
; bx: 61443 -> 57602
mov sp, 998
; sp: 0 -> 998
mov bp, 999
; bp: 0 -> 999
cmp bp, sp
add bp, 1027
; bp: 999 -> 2026
sub bp, 2026
; ZF set
; PF set
; bp: 2026 -> 0
