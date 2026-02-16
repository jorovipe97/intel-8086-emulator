; listing_0047_challenge_flags
bits 16
add bx, 30000
; PF set
; bx: 0 -> 30000
add bx, 10000
; SF set
; OF set
; bx: 30000 -> 40000
sub bx, 5000
; SF set
; PF set
; bx: 40000 -> 35000
sub bx, 5000
; PF set
; OF set
; bx: 35000 -> 30000
mov bx, 1
; bx: 30000 -> 1
mov cx, 100
; cx: 0 -> 100
add bx, cx
; PF set
; bx: 1 -> 101
mov dx, 10
; dx: 0 -> 10
sub cx, dx
; PF set
; cx: 100 -> 90
add bx, -25536
; SF set
; PF set
; bx: 101 -> 40101
add cx, -90
; ZF set
; PF set
; CF set
; cx: 90 -> 0
mov sp, 99
; sp: 0 -> 99
mov bp, 98
; bp: 0 -> 98
cmp bp, sp
; SF set
; PF set
; CF set
