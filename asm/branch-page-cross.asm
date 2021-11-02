  jmp $82fd
  .org $02fd
start:
  ldx  #$02
loop:
  dex
  bne  loop
  jmp  $82fd
