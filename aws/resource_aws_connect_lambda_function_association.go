package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tfconnect "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/waiter"
)

func resourceAwsConnectLambdaFunctionAssociation() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsConnectLambdaFunctionAssociationCreate,
		ReadContext:   resourceAwsConnectLambdaFunctionAssociationRead,
		UpdateContext: resourceAwsConnectLambdaFunctionAssociationRead,
		DeleteContext: resourceAwsConnectLambdaFunctionAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(waiter.ConnectLambdaFunctionAssociationCreatedTimeout),
			Delete: schema.DefaultTimeout(waiter.ConnectInstanceDeletedTimeout),
		},
		Schema: map[string]*schema.Schema{
			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"function_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
		},
	}
}

func resourceAwsConnectLambdaFunctionAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	input := &connect.AssociateLambdaFunctionInput{
		InstanceId:  aws.String(d.Get("instance_id").(string)),
		FunctionArn: aws.String(d.Get("function_arn").(string)),
	}

	lfaid := tfconnect.LambdaFunctionAssociationID(d.Get("instance_id").(string), d.Get("function_arn").(string))

	log.Printf("[DEBUG] Creating Connect Lambda Association %s", input)

	_, err := conn.AssociateLambdaFunctionWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Connect Lambda Function Association (%s): %s", lfaid, err))
	}

	d.SetId(lfaid)
	return resourceAwsConnectLambdaFunctionAssociationRead(ctx, d, meta)
}

func resourceAwsConnectLambdaFunctionAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseID(d.Id())

	associatedFunctionArn, err := finder.LambdaFunctionAssociationByArnWithContext(ctx, conn, instanceID, functionArn)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error finding Lambda Association by Function ARN (%s): %w", functionArn, err))
	}

	if associatedFunctionArn == "" {
		return diag.FromErr(fmt.Errorf("error finding Lambda Association Function ARN (%s): not found", functionArn))
	}

	d.Set("function_arn", functionArn)
	d.Set("instance_id", instanceID)

	return nil
}

func resourceAwsConnectLambdaFunctionAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).connectconn

	instanceID, functionArn, err := tfconnect.LambdaFunctionAssociationParseID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	input := &connect.DisassociateLambdaFunctionInput{
		InstanceId:  aws.String(instanceID),
		FunctionArn: aws.String(functionArn),
	}

	log.Printf("[DEBUG] Deleting Connect Lambda Function Association %s", d.Id())

	_, err = conn.DisassociateLambdaFunction(input)

	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting Connect Lambda Function Association (%s): %w", d.Id(), err))
	}
	return nil
}
