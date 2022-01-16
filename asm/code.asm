PORTB = $6000
PORTA = $6001
DDRB  = $6002
DDRA  = $6003

E  = %10000000
RW = %01000000
RS = %00100000

  .org $0000

reset:
  lda #%11111111  ; ports 0-7 output
  sta DDRB
  lda #%11100000  ; ports 5-7 output
  sta DDRA

  lda #%00111000  ; set 8-bit mode; 2-line display; 5x8 font
  jsr lcd_instruction
  lda #%00001110  ; Display On/Of; Cursor On/Off; Cursor blinking On/Off
  jsr lcd_instruction
  lda #%00000110  ; Set cursor move direction Increment/Decrement; Shift/No Shift of display
  jsr lcd_instruction
  lda #%00000001  ; Clear display
  jsr lcd_instruction

  ldx #0
print_message:
  lda message,x
  beq loop
  jsr print_char
  inx 
  jmp print_message

loop:
  jmp loop

message:
  .asciiz "Hello, world!"

lcd_wait:
  pha
  lda #%00000000  ; ports 0-7 input
  sta DDRB
lcd_busy:
  lda #RW
  sta PORTA
  lda #(RW | E)
  sta PORTA
  lda PORTB
  and #%10000000
  bne lcd_busy
  lda #RW
  sta PORTA
  lda #%11111111  ; ports 0-7 output
  sta DDRB
  pla
  rts

lcd_instruction:
  jsr lcd_wait
  sta PORTB
  lda #0          ; clear RS/RW/E bits
  sta PORTA
  lda #E          ; Set E bit to send instructions 
  sta PORTA
  lda #0          ; Clear RS/RW/E
  sta PORTA
  rts

print_char:
  jsr lcd_wait
  sta PORTB
  lda #RS         ; Set RS; Clear RW/E
  sta PORTA
  lda #(RS | E)   ; Set E bit to send data
  sta PORTA
  lda #RS         ; Set RS; clear RW/E
  sta PORTA
  rts

;  .org $fffc
;  .word reset
;  .word $0000
