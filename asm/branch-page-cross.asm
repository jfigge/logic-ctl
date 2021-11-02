  jmp $82f5
  .org $02f5
  ldy  #$02
  dey
  beq next
  jmp $82f7
  .org $02fd
  ldx  #$02
loop:
  dex
  bne  loop
next:
  jmp  $82f5
