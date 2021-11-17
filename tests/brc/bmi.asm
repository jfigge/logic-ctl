  .org $0000
  .org $0200
start:
  ldx #$00
  bmi start
  jmp middle

  .org $02b0
middle:
  ldx #$80
  bmi edge
  jmp $0300

  .org $02fd
edge:
  bmi forward
  jmp $0300

forward:
  bmi middle

