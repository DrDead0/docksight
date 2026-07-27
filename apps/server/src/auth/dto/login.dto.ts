import { ApiProperty } from '@nestjs/swagger';
import { IsEmail, IsNotEmpty, IsString } from 'class-validator';

/**
 * Login input. Validation here is only about shape — never about whether the
 * account exists, so an invalid-format email and an unknown account produce
 * different codes (400 vs 401) but neither confirms a registration.
 */
export class LoginDto {
  @ApiProperty({ example: 'admin@example.com' })
  @IsEmail({}, { message: 'A valid email is required' })
  email!: string;

  @ApiProperty({ example: 'super-secret-password' })
  @IsString()
  @IsNotEmpty({ message: 'Password is required' })
  password!: string;
}
