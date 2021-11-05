  lda #$55
  pha
  lda #$ff
  pla
  ldx #$f0
  txs
  ldx #$00
  tsx
  ADC #$ab
  PHP
  LDX #$FF
  CLC
  PLP
  jmp $8000
