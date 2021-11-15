  .org $0000
  .org $0200
  ldx #$10
  asl $0210,x
  asl $0210,x
  asl $0210,x
  inx
  asl $0210,x
  asl $0210,x
  asl $0210,x

  .org $0220
  .byte $24
  .byte $a2
