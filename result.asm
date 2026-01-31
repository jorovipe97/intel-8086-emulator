; listing_0041_add_sub_cmp_jnz
bits 16
add bx, [bx+si]
add bx, [bp]
add si, 2
add bp, 2
add cx, 8
add bx, [bp]
add cx, [bx+2]
add bh, [bp+si+4]
add di, [bp+di+6]
add word [bx+si], bx
add word [bp], bx
add word [bp], bx
add word [bx+2], cx
add byte [bp+si+4], bh
add word [bp+di+6], di
add byte [bx], 34
add word [bp+si+1000], 29
add ax, [bp]
add al, [bx+si]
add ax, bx
add al, ah
add ax, 1000
add al, -30
add al, 9
sub bx, [bx+si]
sub bx, [bp]
sub si, 2
sub bp, 2
sub cx, 8
sub bx, [bp]
sub cx, [bx+2]
sub bh, [bp+si+4]
sub di, [bp+di+6]
sub word [bx+si], bx
sub word [bp], bx
sub word [bp], bx
sub word [bx+2], cx
sub byte [bp+si+4], bh
sub word [bp+di+6], di
sub byte [bx], 34
sub word [bx+di], 29
sub ax, [bp]
sub al, [bx+si]
sub ax, bx
sub al, ah
sub ax, 1000
sub al, -30
sub al, 9
; ERROR: Unrecognized binary in instruction stream.
