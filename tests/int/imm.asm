  .org $0000

  .org $0200
  ldx #$00
  inx
  jmp $0202

  .org $0230
  ldx #$00
  rti
;
;  .org $fffe
;  .word $0230
